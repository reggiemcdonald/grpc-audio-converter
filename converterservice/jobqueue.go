// A job queue for file conversion requests
// Created using the tutorials at https://gist.github.com/harlow/dbcd639cf8d396a2ab73
// and https://riptutorial.com/go/example/18325/job-queue-with-worker-pool
package converterservice

import (
	"errors"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/fileconverter"
	"sync"
)

type FileConverterJobQueue interface {
	Enqueue(request FileConverterJob) error
	Start() error
	Stop()
	Running() bool
}

type FileConverterJob interface {
	Start()
}

type JobQueueConfiguration struct {
	Concurrency int
	QueueSize int
}

type worker struct {
	waitGroup *sync.WaitGroup
	freeChans chan chan FileConverterJob
	thisChan chan FileConverterJob
	stopWorker chan bool
}

type jobQueue struct {
	converter fileconverter.Converter
	// A channel of channels that contains worker threads
	readyWorkers chan chan FileConverterJob
	readyJobs chan FileConverterJob
	running bool
	stopAll chan bool
	workers []*worker
}

func newWorker(waitGroup *sync.WaitGroup, freeChans chan chan FileConverterJob) *worker {
	return &worker{
		waitGroup: waitGroup,
		freeChans: freeChans,
		thisChan: make(chan FileConverterJob),
		stopWorker: make(chan bool),
	}
}

func NewJobQueue(config *JobQueueConfiguration) FileConverterJobQueue {
	waitGroup := &sync.WaitGroup{}
	readyWorkers := make(chan chan FileConverterJob, config.Concurrency)
	readyJobs := make(chan FileConverterJob, config.QueueSize)
	workers := make([]*worker, config.Concurrency)
	for i, _ := range workers {
		workers[i] = newWorker(waitGroup, readyWorkers)
	}
	return &jobQueue{
		readyWorkers: readyWorkers,
		readyJobs: readyJobs,
		running: false,
		stopAll: make(chan bool),
		workers: workers,
	}
}

func (q *jobQueue) Enqueue(job FileConverterJob) error {
	if !q.running {
		return errors.New("queue is shutdown")
	}
	select {
	case q.readyJobs <- job:
		return nil
	default:
		return errors.New("too many requests")
	}
}

func (q *jobQueue) Start() error {
	for _, w := range q.workers {
		if err := w.start(); err != nil {
			return errors.New("encountered problem starting job queue")
		}
	}
	go q.run()
	q.running = true
	return nil
}

func (q *jobQueue) Stop() {
	q.running = false
	for _, w := range q.workers {
		w.stop()
	}
}

func (q *jobQueue) Running() bool {
	return q.running
}

func (q *jobQueue) run() {
	for {
		select {
		case newJob := <- q.readyJobs:
			availableWorkerChannel := <- q.readyWorkers
			availableWorkerChannel <- newJob
		case <- q.stopAll:
			for _, w := range q.workers {
				w.stop()
			}
		}
	}
}

func (w *worker) start() error {
	go w.run()
	return nil
}

func (w *worker) stop() {
	w.stopWorker <- true
}

func (w *worker) run() {
	w.waitGroup.Add(1)
	for {
		w.freeChans <- w.thisChan
		select {
		case job := <- w.thisChan:
			job.Start()
		case <- w.stopWorker:
			w.waitGroup.Done()
			return
		}
	}
}

