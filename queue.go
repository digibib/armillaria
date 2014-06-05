package main

import (
	"encoding/json"
	"fmt"

	"github.com/digibib/armillaria/sparql"
	log "gopkg.in/inconshreveable/log15.v2"
)

const resourceQuery = "SELECT * WHERE { %s ?p ?o }"

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
	Who() int
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
			r, err := db.Query(fmt.Sprintf(resourceQuery, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, uri)})
				break
			}

			var res *sparql.Results
			err = json.Unmarshal(r, &res)
			if err != nil {
				l.Error("failed to parse sparql response", log.Ctx{"error": err.Error(), "uri": uri})
				break
			}

			// fetch the resource profile from the SPARQL response
			var profile string
			published := false
			for _, b := range res.Results.Bindings {
				if b["p"].Value == "armillaria://internal/profile" {
					profile = b["o"].Value
				}
				if b["p"].Value == "armillaria://internal/published" {
					// we need to know which index to write to
					published = true
				}
			}
			if profile == "" {
				l.Error("resource lacks profile information", log.Ctx{"uri": uri})
				break
			}

			resource := make(map[string]string)
			var pred string
			for _, b := range res.Results.Bindings {
				pred = urlify(b["p"].Value)
				if indexMappings[profile][pred] == "" {
					continue // if not in mapping, we don't want to index it
				}
				resource[indexMappings[profile][pred]] = b["o"].Value
			}

			// We want to use the URI as the elasticsearch document ID
			resource["uri"] = string(uri[1 : len(uri)-1])
			resourceBody, err := json.Marshal(resource)
			if err != nil {
				l.Error("failed to marshal json", log.Ctx{"error": err.Error(), "uri": uri})
			}

			var index = "drafts"
			if published {
				index = "public"
			}

			err = esIndexer.Add(index, profile, resourceBody)
			if err != nil {
				log.Error("failed to index resource", log.Ctx{"error": err.Error(), "uri": uri})
				break
			}

			l.Info("indexed resource", log.Ctx{"uri": uri, "workerID": w.Who(), "index": "public", "profile": profile})
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
