package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/digibib/armillaria/sparql"
	log "gopkg.in/inconshreveable/log15.v2"
)

const resourceQuery = `
SELECT *
FROM <%s>
WHERE {
   { %s ?p ?o .
     MINUS { %s ?p ?o . ?o <armillaria://internal/displayLabel> _:l . } }
   UNION
   { %s ?p ?o .
     ?o <armillaria://internal/displayLabel> ?l . }
}`

const affectedResourcesQuery = `
SELECT ?resource
FROM <%s>
WHERE {
	{ ?resource _:p %s } UNION { %s _:p ?resource }
	?resource <armillaria://internal/profile> "manifestation" .
} LIMIT 100
`

const insertKohaIDQuery = `
INSERT DATA
 { GRAPH <%s>
	{ %v <armillaria://internal/kohaID> %v } }`

// indexRequest holds the URI which should be indexed or removed from an index.
type indexRequest string

// workerFactory is the function signature for creating a worker.
type workerFactory func(int, chan chan indexRequest) Worker

func urlify(s string) string { return fmt.Sprintf("<%s>", s) }

// Queue is an in-memory work queue.
type Queue struct {
	Name          string
	NumWorkers    int
	WorkQueue     chan indexRequest
	WorkerQueue   chan chan indexRequest
	WorkerFactory workerFactory
}

func (q Queue) runDispatcher() {
	q.WorkerQueue = make(chan chan indexRequest, q.NumWorkers)

	for i := 0; i < q.NumWorkers; i++ {
		w := q.WorkerFactory(i+1, q.WorkerQueue)
		go w.Run()
		l.Info("staring worker", log.Ctx{"queue": q.Name, "workerID": w.ID()})
	}

	for {
		select {
		case work := <-q.WorkQueue:
			go func() {
				worker := <-q.WorkerQueue
				worker <- work
			}()
		}
	}
}

func newQueue(name string, bufferSize int, numWorkers int, wFn workerFactory) Queue {
	return Queue{
		Name:          name,
		WorkQueue:     make(chan indexRequest, bufferSize),
		NumWorkers:    numWorkers,
		WorkerFactory: wFn,
	}
}

// Worker is the interface witch all queue workers must implement.
type Worker interface {
	ID() int // TODO rename ID()
	Run()
	Stop()
}

type addWorker struct {
	id          int
	work        chan indexRequest
	workerQueue chan chan indexRequest
	quit        chan bool
}

func (w addWorker) Run() {
	for {
		// ready for new work
		w.workerQueue <- w.work

		select {
		case uri := <-w.work:
			// Get RDF resource to be indexed
			r, err := db.Query(fmt.Sprintf(resourceQuery, cfg.RDFStore.DefaultGraph, uri, uri, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, uri, uri, uri)})
				// TODO uri should be stored for retry
				break
			}

			// Generate JSON document to be sent to Elasticsearch
			resourceBody, profile, err := createIndexDoc(indexMappings, r, string(uri[1:len(uri)-1]))
			if err != nil {
				l.Error("failed to create indexable json doc from RDF resource", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			// Index document
			err = esIndexer.Add("public", profile, resourceBody)
			if err != nil {
				log.Error("failed to index resource", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			l.Info("indexed resource", log.Ctx{"uri": uri, "workerID": w.ID(), "index": "public", "profile": profile})

			// Send uri for sync to Koha
			queueKohaSync.WorkQueue <- uri

			// Check if there are other resources which are affected by this resource.
			r, err = db.Query(fmt.Sprintf(affectedResourcesQuery, cfg.RDFStore.DefaultGraph, uri, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, uri, uri, uri)})
				// TODO uri should be stored for retry
				break
			}
			var res *sparql.Results

			err = json.Unmarshal(r, &res)
			if err != nil {
				l.Error("failed to parse sparql response", log.Ctx{"error": err.Error(), "uri": uri})
			}
			for _, b := range res.Results.Bindings {
				queueKohaSync.WorkQueue <- indexRequest("<" + b["resource"].Value + ">")
			}

		case <-w.quit:
			println(w.ID(), "quitting")
			return
		}
	}
}

func (w addWorker) ID() int {
	return w.id
}

func (w addWorker) Stop() {
	w.quit <- true
}

func newAddWorker(id int, wq chan chan indexRequest) Worker {
	return addWorker{
		id:          id,
		work:        make(chan indexRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}

type rmWorker struct {
	id          int
	work        chan indexRequest
	workerQueue chan chan indexRequest
	quit        chan bool
}

func (w rmWorker) Run() {
	for {
		w.workerQueue <- w.work

		select {
		case uri := <-w.work:
			err := esIndexer.Remove(string(uri[1 : len(uri)-1]))
			if err != nil {
				log.Error("failed to remove resource from index", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			l.Info("removed resource from index", log.Ctx{"uri": uri, "workerID": w.ID()})
		case <-w.quit:
			println(w.ID(), "quitting")
			return
		}
	}
}

func (w rmWorker) ID() int {
	return w.id
}

func (w rmWorker) Stop() {
	w.quit <- true
}

func newRmWorker(id int, wq chan chan indexRequest) Worker {
	return rmWorker{
		id:          id,
		work:        make(chan indexRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}

type kohaSyncWorker struct {
	id          int
	work        chan indexRequest
	workerQueue chan chan indexRequest
	quit        chan bool
}

func (w kohaSyncWorker) Run() {
	for {
		w.workerQueue <- w.work

		select {
		case uri := <-w.work:
			// Get RDF of resource
			// TODO should be same query as needed for RDF2MARC? or just a slim response with armillaria properties?
			r, err := db.Query(fmt.Sprintf(resourceQuery, cfg.RDFStore.DefaultGraph, uri, uri, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, uri, uri, uri)})
				// TODO uri should be stored for retry
				break
			}
			var res *sparql.Results
			var profile string
			err = json.Unmarshal(r, &res)
			if err != nil {
				l.Error("failed to parse sparql response", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			if len(res.Results.Bindings) == 0 {
				l.Error("cannot sync non-existing resource to Koha", log.Ctx{"error": err.Error(), "uri": uri})
				break
			}

			var bibnrStr string
			for _, b := range res.Results.Bindings {
				if b["p"].Value == "armillaria://internal/profile" {
					profile = b["o"].Value
				}
				if b["p"].Value == "armillaria://internal/kohaID" {
					bibnrStr = b["o"].Value
				}
			}

			// We are only syncing manifestations to Koha
			if profile != "manifestation" {
				break
			}

			// Make sure we are authenticated to Koha
			if kohaCookies == nil {
				kohaCookies, err = syncKohaAuth(cfg.KohaPath, cfg.KohaSyncUser, cfg.KohaSyncPass)
				if err != nil {
					l.Error("cannot authenticate to Koha /svc API", log.Ctx{"error": err.Error(), "uri": uri})
					// TODO uri should be stored for retry
					break
				}
			}

			// Generate MARCXML record of RDF resource
			r, err = db.Query(fmt.Sprintf(queryRDF2MARC, cfg.RDFStore.DefaultGraph, uri, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error()})
				// TODO uri should be stored for retry
				break
			}

			err = json.Unmarshal(r, &res)
			if err != nil {
				l.Error("failed to parse sparql response", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			rec, err := convertRDF2MARC(*res)
			if err != nil {
				l.Error("failed to generate marc record from RDF", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			marc, err := xml.Marshal(rec)
			if err != nil {
				l.Error("failed to marshal marc record into XML", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			var bibnr int
			// If resorurce has a kohaID, its an update, otherwise it's a new resource
			if bibnrStr != "" {
				// we're updating
				bibnr, err = strconv.Atoi(bibnrStr)
				if err != nil {
					l.Error("kohaID on resource is not an integer", log.Ctx{"error": err.Error(), "uri": uri})
					// TODO uri should be stored for retry
					break
				}

				err = syncUpdatedManifestation(cfg.KohaPath, kohaCookies, marc, bibnr)
				if err != nil {
					l.Error("sync updated resource to Koha failed", log.Ctx{"error": err.Error(), "uri": uri})
					// TODO uri should be stored for retry
					break
				}
			} else {
				// uri is a new resource
				bibnr, err = syncNewManifestation(cfg.KohaPath, kohaCookies, marc)
				if err != nil {
					l.Error("sync new resource to Koha failed", log.Ctx{"error": err.Error(), "uri": uri})
					// TODO uri should be stored for retry
					break
				}

				// store the koha id as property on the RDF resource
				r, err := db.Query(fmt.Sprintf(insertKohaIDQuery, cfg.RDFStore.DefaultGraph, uri, bibnr))
				if err != nil {
					l.Error("db.Query failed", log.Ctx{"error": err.Error()})
					// TODO uri should be stored for retry
					break
				}

				if bytes.Index(r, []byte("1 (or less) triples")) == -1 {
					l.Error("failed to insert koha ID on resource", log.Ctx{"error": err.Error(),
						"sparqlResponse": string(r)})
					// TODO uri should be stored for retry
					break
				}
			}

			l.Info("synced resource to Koha", log.Ctx{"uri": uri, "workerID": w.ID(), "biblionr": bibnr})
		case <-w.quit:
			println(w.ID(), "quitting")
			return
		}
	}
}

func (w kohaSyncWorker) ID() int {
	return w.id
}

func (w kohaSyncWorker) Stop() {
	w.quit <- true
}

func newKohaSyncWorker(id int, wq chan chan indexRequest) Worker {
	return kohaSyncWorker{
		id:          id,
		work:        make(chan indexRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}
