package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/digibib/armillaria/sparql"
	log "gopkg.in/inconshreveable/log15.v2"
)

// How many times to try to sync to Koha before we give up
const maxRetries = 3

// qRequest holds the URI/biblinr which should be indexed and synced,
// and keeps track of how many times the sync has been attemtped.
type qRequest struct {
	uri      string
	biblionr int
	task     string // createDraft|updateDraft|deleteDraft|create|update|delete
	count    int
}

// workerFactory is the function signature for creating a worker.
type workerFactory func(int, chan chan qRequest) Worker

func urlify(s string) string { return fmt.Sprintf("<%s>", s) }

func retry(job qRequest, c chan qRequest, queue string) {
	job.count = job.count + 1
	if job.count >= maxRetries {
		l.Info("gave up resource after max retries", log.Ctx{"uri": job.uri, "task": job.task, "queue": queue})
		return
	}

	time.Sleep(time.Second * time.Duration(job.count*10))
	c <- job
}

// Queue is an in-memory work queue.
type Queue struct {
	Name          string
	NumWorkers    int
	WorkQueue     chan qRequest
	WorkerQueue   chan chan qRequest
	WorkerFactory workerFactory
	ShutDown      chan bool
	Workers       []Worker
}

type Queues []Queue

func (qs Queues) Get(name string) (*Queue, error) {
	for _, q := range qs {
		if q.Name == name {
			return &q, nil
		}
	}
	return nil, fmt.Errorf("found no queue with name: %s", name)
}

func (qs Queues) StartAll() {
	for _, q := range qs {
		go q.runDispatcher()
	}
}

func (qs Queues) StopAll() {
	for _, q := range qs {
		q.ShutDown <- true
		<-q.ShutDown
	}
}

func (q Queue) runDispatcher() {
	q.WorkerQueue = make(chan chan qRequest, q.NumWorkers)

	for i := 0; i < q.NumWorkers; i++ {
		w := q.WorkerFactory(i+1, q.WorkerQueue)
		go w.Run()
		l.Info("staring worker", log.Ctx{"queue": q.Name, "workerID": w.ID()})
		q.Workers = append(q.Workers, w)
	}

	for {
		select {
		case job := <-q.WorkQueue:
			go func() {
				worker := <-q.WorkerQueue
				worker <- job
			}()
		case <-q.ShutDown:
			for _, w := range q.Workers {
				l.Info("stopping worker", log.Ctx{"queue": q.Name, "workerID": w.ID()})
				w.Stop()
			}
			q.ShutDown <- true
		}
	}
}

func newQueue(name string, bufferSize int, numWorkers int, wFn workerFactory) Queue {
	return Queue{
		Name:          name,
		WorkQueue:     make(chan qRequest, bufferSize),
		NumWorkers:    numWorkers,
		WorkerFactory: wFn,
		ShutDown:      make(chan bool),
	}
}

// Worker is the interface for queue workers; they should
// know how to start, stop and identify themselves.
type Worker interface {
	ID() int
	Run()
	Stop()
}

type addWorker struct {
	id          int
	work        chan qRequest
	workerQueue chan chan qRequest
	quit        chan bool
}

func (w addWorker) Run() {
	for {
		// ready for new work
		w.workerQueue <- w.work

		select {
		case job := <-w.work:
			// Get RDF resource to be indexed
			var graph = cfg.RDFStore.DefaultGraph
			if job.task == "createDraft" || job.task == "updateDraft" {
				graph = cfg.RDFStore.DraftsGraph
			}
			r, err := db.Query(fmt.Sprintf(resourceQuery, graph, job.uri, job.uri, job.uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error()})
				if q, err := queues.Get("add"); err == nil {
					retry(job, q.WorkQueue, q.Name)
				}
				break
			}

			// Generate JSON document to be sent to Elasticsearch
			resourceBody, profile, err := createIndexDoc(indexMappings, r, string(job.uri[1:len(job.uri)-1]))
			if err != nil {
				l.Error("failed to create indexable json doc from RDF resource", log.Ctx{"error": err.Error(), "uri": job.uri})
				if q, err := queues.Get("add"); err == nil {
					retry(job, q.WorkQueue, q.Name)
				}
				break
			}

			// Index document
			err = esIndexer.Add("public", profile, resourceBody)
			if err != nil {
				log.Error("failed to index resource", log.Ctx{"error": err.Error(), "uri": job.uri})
				if q, err := queues.Get("add"); err == nil {
					retry(job, q.WorkQueue, q.Name)
				}
				break
			}

			l.Info("indexed resource", log.Ctx{"uri": job.uri, "workerID": w.ID(), "index": "public", "profile": profile})

			// Send uri for sync to Koha
			if cfg.SyncToKoha {
				if q, err := queues.Get("KohaSync"); err == nil {
					q.WorkQueue <- job
				}
			}

			// Check if there are other resources which are affected by this resource.
			r, err = db.Query(fmt.Sprintf(affectedResourcesQuery, cfg.RDFStore.DefaultGraph, job.uri, job.uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, job.uri, job.uri, job.uri)})
				if q, err := queues.Get("add"); err == nil {
					retry(job, q.WorkQueue, q.Name)
				}
				break
			}
			var res *sparql.Results

			err = json.Unmarshal(r, &res)
			if err != nil {
				l.Error("failed to parse sparql response", log.Ctx{"error": err.Error(), "uri": job.uri})
				// TODO now what, retry?
				break
			}
			for _, b := range res.Results.Bindings {
				if q, err := queues.Get("KohaSync"); err == nil {
					job := qRequest{uri: "<" + b["resource"].Value + ">"}
					if biblionr, err := strconv.Atoi(b["kohaID"].Value); err == nil {
						job.biblionr = biblionr
					}
					q.WorkQueue <- job
				}
			}

		case <-w.quit:
			return
		}
	}
}

func (w addWorker) ID() int { return w.id }
func (w addWorker) Stop()   { w.quit <- true }

func newAddWorker(id int, wq chan chan qRequest) Worker {
	return addWorker{
		id:          id,
		work:        make(chan qRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}

type rmWorker struct {
	id          int
	work        chan qRequest
	workerQueue chan chan qRequest
	quit        chan bool
}

func (w rmWorker) Run() {
	for {
		w.workerQueue <- w.work

		select {
		case job := <-w.work:
			err := esIndexer.Remove(string(job.uri[1 : len(job.uri)-1]))
			if err != nil {
				log.Error("failed to remove resource from index", log.Ctx{"error": err.Error(), "uri": job.uri})
				if q, err := queues.Get("remove"); err == nil {
					retry(job, q.WorkQueue, q.Name)
				}
				break
			}

			// send to Koha-sync queue for deletion, if not a draf
			if job.task != "deleteDraft" && cfg.SyncToKoha {
				if q, err := queues.Get("KohaSync"); err == nil {
					q.WorkQueue <- job
				}
			}

			l.Info("removed resource from index", log.Ctx{"uri": job.uri, "workerID": w.ID()})
		case <-w.quit:
			return
		}
	}
}

func (w rmWorker) ID() int { return w.id }
func (w rmWorker) Stop()   { w.quit <- true }

func newRmWorker(id int, wq chan chan qRequest) Worker {
	return rmWorker{
		id:          id,
		work:        make(chan qRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}

type kohaSyncWorker struct {
	id          int
	work        chan qRequest
	workerQueue chan chan qRequest
	quit        chan bool
}

func (w kohaSyncWorker) Run() {
	for {
		w.workerQueue <- w.work

		select {
		case job := <-w.work:
			var err error
			var tryAgain bool
			var biblionr int

			switch job.task {
			case "create":
				biblionr, tryAgain, err = syncCreateResource(job.uri)
			case "update":
				tryAgain, err = syncUpdateResource(job.uri, job.biblionr)
			case "delete":
				tryAgain, err = syncDeleteResource(job.biblionr)
			default:
				continue
			}
			if err != nil {
				l.Error("sync resource to Koha failed",
					log.Ctx{"error": err.Error(), "uri": job.uri, "worker": w.ID(), "task": job.task})
				if tryAgain {
					if q, err := queues.Get("KohaSync"); err == nil {
						retry(job, q.WorkQueue, q.Name)
					}
					continue
				}
			}

			if biblionr == 0 {
				biblionr = job.biblionr
			}

			l.Info("synced resource to Koha",
				log.Ctx{"uri": job.uri, "workerID": w.ID(), "biblionr": biblionr, "task": job.task})

		case <-w.quit:
			return
		}
	}
}

func (w kohaSyncWorker) ID() int { return w.id }
func (w kohaSyncWorker) Stop()   { w.quit <- true }

func newKohaSyncWorker(id int, wq chan chan qRequest) Worker {
	return kohaSyncWorker{
		id:          id,
		work:        make(chan qRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}
