package main

import (
	"encoding/json"
	"fmt"

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
}
`

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
		l.Info("staring worker", log.Ctx{"queue": q.Name, "workerID": w.Who()})
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
	Who() int // TODO rename ID()
	Run()
	Stop()
}

type addWorker struct {
	ID          int
	Work        chan indexRequest
	WorkerQueue chan chan indexRequest
	Quit        chan bool
}

func (w addWorker) Run() {
	for {
		// ready for new work
		w.WorkerQueue <- w.Work

		select {
		case uri := <-w.Work:
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

			l.Info("indexed resource", log.Ctx{"uri": uri, "workerID": w.Who(), "index": "public", "profile": profile})

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

		case <-w.Quit:
			println(w.Who(), "quitting")
			return
		}
	}
}

func (w addWorker) Who() int {
	return w.ID
}

func (w addWorker) Stop() {
	w.Quit <- true
}

func newAddWorker(id int, wq chan chan indexRequest) Worker {
	return addWorker{
		ID:          id,
		Work:        make(chan indexRequest),
		WorkerQueue: wq,
		Quit:        make(chan bool, 1),
	}
}

type rmWorker struct {
	ID          int
	Work        chan indexRequest
	WorkerQueue chan chan indexRequest
	Quit        chan bool
}

func (w rmWorker) Run() {
	for {
		w.WorkerQueue <- w.Work

		select {
		case uri := <-w.Work:
			err := esIndexer.Remove(string(uri[1 : len(uri)-1]))
			if err != nil {
				log.Error("failed to remove resource from index", log.Ctx{"error": err.Error(), "uri": uri})
				// TODO uri should be stored for retry
				break
			}

			l.Info("removed resource from index", log.Ctx{"uri": uri, "workerID": w.Who()})
		case <-w.Quit:
			println(w.Who(), "quitting")
			return
		}
	}
}

func (w rmWorker) Who() int {
	return w.ID
}

func (w rmWorker) Stop() {
	w.Quit <- true
}

func newRmWorker(id int, wq chan chan indexRequest) Worker {
	return rmWorker{
		ID:          id,
		Work:        make(chan indexRequest),
		WorkerQueue: wq,
		Quit:        make(chan bool, 1),
	}
}

type kohaSyncWorker struct {
	ID          int
	Work        chan indexRequest
	WorkerQueue chan chan indexRequest
	Quit        chan bool
}

func (w kohaSyncWorker) Run() {
	for {
		w.WorkerQueue <- w.Work

		select {
		case uri := <-w.Work:
			// Get RDF of resource
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
				l.Error("cannot sync non-existing resource to Koha", log.Ctx{"error": err.Error(), "uri": uri}
				break
			}

			for _, b := range res.Results.Bindings {
				if b["p"].Value == "armillaria://internal/profile" {
					profile = b["o"].Value
					break
				}
			}

			if profile != "manifestation" {
				// We are only syncing manifestations to Koha
				break
			}

			l.Info("synced resource to Koha (simulated)", log.Ctx{"uri": uri, "workerID": w.Who()})
		case <-w.Quit:
			println(w.Who(), "quitting")
			return
		}
	}
}

func (w kohaSyncWorker) Who() int {
	return w.ID
}

func (w kohaSyncWorker) Stop() {
	w.Quit <- true
}

func newKohaSyncWorker(id int, wq chan chan indexRequest) Worker {
	return kohaSyncWorker{
		ID:          id,
		Work:        make(chan indexRequest),
		WorkerQueue: wq,
		Quit:        make(chan bool, 1),
	}
}
