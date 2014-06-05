package main

import (
	log "gopkg.in/inconshreveable/log15.v2"
)

// indexRequest holds the URI which should be indexed or removed from an index.
type indexRequest string

// workerFactory is the function signature for creating a worker.
type workerFactory func(int, chan chan indexRequest) Worker

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
		l.Info("staring worker", log.Ctx{"queue": q.Name, "id2": w.Who()})
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
			println(w.Who(), "adding to index:", uri)
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
			println(w.Who(), "removing from index:", uri)
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
