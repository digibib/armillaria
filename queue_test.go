package main

import (
	"reflect"
	"testing"
)

type testJob struct {
	handeledBy int
	uri        string
}

var answers = make(chan testJob)

type testWorker struct {
	id          int
	work        chan qRequest
	workerQueue chan chan qRequest
	quit        chan bool
}

func (w testWorker) Run() {
	for {
		w.workerQueue <- w.work

		select {
		case job := <-w.work:
			answers <- testJob{handeledBy: w.ID(), uri: job.uri}
		case <-w.quit:
			return
		}
	}
}

func (w testWorker) ID() int { return w.id }
func (w testWorker) Stop()   { w.quit <- true }

func newTestWorker(id int, wq chan chan qRequest) Worker {
	return testWorker{
		id:          id,
		work:        make(chan qRequest),
		workerQueue: wq,
		quit:        make(chan bool, 1),
	}
}

func TestQueue(t *testing.T) {
	tq := newQueue("testQueue", 10, 5, newTestWorker)
	go tq.runDispatcher()

	tq.WorkQueue <- qRequest{uri: "<a>"}
	tq.WorkQueue <- qRequest{uri: "<b>"}

	got := <-answers
	want := testJob{handeledBy: 1, uri: "<a>"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	got = <-answers
	want = testJob{handeledBy: 2, uri: "<b>"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	tq.ShutDown <- true
}
