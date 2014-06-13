package main

import (
	"encoding/json"
	"fmt"

	"github.com/digibib/armillaria/sparql"
	log "gopkg.in/inconshreveable/log15.v2"
)

const resourceQuery = `
SELECT * WHERE {
   { %s ?p ?o .
     MINUS { %s ?p ?o . ?o <armillaria://internal/displayLabel> _:l . } }
   UNION
   { %s ?p ?o .
     ?o <armillaria://internal/displayLabel> ?l . }
}`

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
			r, err := db.Query(fmt.Sprintf(resourceQuery, uri, uri, uri))
			if err != nil {
				l.Error("db.Query failed", log.Ctx{"error": err.Error(), "query": fmt.Sprintf(resourceQuery, uri, uri, uri)})
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
			for _, b := range res.Results.Bindings {
				if b["p"].Value == "armillaria://internal/profile" {
					profile = b["o"].Value
					break
				}
			}
			if profile == "" {
				l.Error("resource lacks profile information", log.Ctx{"uri": uri})
				break
			}

			resource := make(map[string]interface{})
			type uriField struct {
				URI   string `json:"uri"`
				Label string `json:"label"`
			}
			var pred string
			uf := uriField{}

			for _, b := range res.Results.Bindings {
				pred = urlify(b["p"].Value)
				if indexMappings[profile][pred] == "" {
					continue // if not in mapping, we don't want to index it
				}
				if _, ok := b["l"]; ok {
					uf.URI = b["o"].Value
					uf.Label = b["l"].Value
					switch resource[indexMappings[profile][pred]].(type) {
					case []interface{}:
						resource[indexMappings[profile][pred]] =
							append(resource[indexMappings[profile][pred]].([]interface{}), uf)
					case uriField:
						var s []interface{}
						s = append(s, resource[indexMappings[profile][pred]])
						resource[indexMappings[profile][pred]] = append(s, uf)
					default:
						resource[indexMappings[profile][pred]] = uf
					}
					continue
				}

				val := b["o"].Value
				switch resource[indexMappings[profile][pred]].(type) {
				case []interface{}:
					resource[indexMappings[profile][pred]] =
						append(resource[indexMappings[profile][pred]].([]interface{}), val)
				case interface{}:
					var s []interface{}
					s = append(s, resource[indexMappings[profile][pred]])
					resource[indexMappings[profile][pred]] = append(s, val)
				default:
					resource[indexMappings[profile][pred]] = val
				}

			}

			// We want to use the URI as the elasticsearch document ID
			resource["uri"] = string(uri[1 : len(uri)-1])

			resourceBody, err := json.Marshal(resource)
			if err != nil {
				l.Error("failed to marshal json", log.Ctx{"error": err.Error(), "uri": uri})
			}

			err = esIndexer.Add("public", profile, resourceBody)
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
			err := esIndexer.Remove(string(uri[1 : len(uri)-1]))
			if err != nil {
				log.Error("failed to remove resource from index", log.Ctx{"error": err.Error(), "uri": uri})
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
