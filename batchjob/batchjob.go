package batchjob

import (
	"sync"
	"time"
)

// JobResult
type JobResult struct {
	Job     Job
	Result  interface{}
	Error   error
	Success bool
}

// Job interface
type Job interface {
	Execute() (interface{}, error)
}

// endJob is a special type of job that indicates that all jobs in a batch were processed
type endJob struct {
}

func (e *endJob) Execute() (interface{}, error) {
	return "nil", nil
}

// BatchProcessor

type state bool

const (
	Active   state = true
	Inactive state = false
)

type batchProcessorInner struct {
	jobCache         []Job
	batchSize        uint16
	callback         *func([]JobResult)
	batchInterval    time.Duration
	state            state
	beginFrom        int
	batchesInProcess uint
}

type BatchProccessInitialiser struct {
	BatchSize uint16
	Interval  time.Duration
	Callback  *func([]JobResult)
}

type BatchProcessor batchProcessorInner

func InstantiateBatchProcessor(initialiser BatchProccessInitialiser) (*BatchProcessor, bool) {
	if initialiser.Callback == nil || (initialiser.BatchSize == 0 && initialiser.Interval == 0) {
		return nil, false
	}
	return &BatchProcessor{jobCache: []Job{}, batchSize: initialiser.BatchSize, batchInterval: initialiser.Interval, callback: initialiser.Callback}, true
}

func (bp *BatchProcessor) AddJob(job Job) {
	bp.jobCache = append(bp.jobCache, job)
	if bp.state == Active && bp.batchesInProcess < 1 {
		println("Restarting Process")
		go bp.Begin()
		time.Sleep(5000 * time.Millisecond)
	}
}

func (bp *BatchProcessor) RemoveJob(job Job) bool {
	newJobCache := []Job{}
	var removed = false
	for _, j := range bp.jobCache {
		if j == job {
			removed = true
		} else {
			newJobCache = append(newJobCache, j)
		}
	}
	bp.jobCache = newJobCache
	return removed
}

func (bp *BatchProcessor) Count() int {
	return len(bp.jobCache)
}

func (bp *BatchProcessor) processBatch(batch []Job, waiter *sync.WaitGroup) {
	bp.batchesInProcess = bp.batchesInProcess + 1
	waiter.Add(1)
	results := []JobResult{}
	for _, job := range batch {
		result, err := job.Execute()
		results = append(results, JobResult{Job: job, Result: result, Error: err, Success: err == nil})
	}
	callback := *bp.callback
	callback(results)
	waiter.Done()
	bp.batchesInProcess = bp.batchesInProcess - 1
}

func (bp *BatchProcessor) getNextBatch() []Job {
	if len(bp.jobCache) == 0 {
		return []Job{}
	}

	var sliceStart = bp.beginFrom
	var sliceEnd int
	if bp.batchSize == 0 {
		sliceEnd = len(bp.jobCache)
	} else {
		sliceEnd = sliceStart + int(bp.batchSize)
	}
	if sliceEnd > len(bp.jobCache) {
		sliceEnd = len(bp.jobCache)
	}
	bp.beginFrom = sliceEnd

	return bp.jobCache[sliceStart:sliceEnd]
}

func (bp *BatchProcessor) atEnd() bool {
	return bp.beginFrom >= len(bp.jobCache)
}

func (bp *BatchProcessor) Begin() chan bool {
	println("BEGIN")
	var asyncProcessWait sync.WaitGroup
	// defer asyncProcessWait.Wait()

	processListener := make(chan bool)

	bp.state = Active

	if bp.batchInterval != 0 {
	batchLoop:
		for {
			var alarm = <-time.Tick(bp.batchInterval)
			switch alarm {
			default:
				batch := bp.getNextBatch()
				go func(batch []Job, waiter *sync.WaitGroup) {
					bp.processBatch(batch, waiter)
				}(batch, &asyncProcessWait)
				if bp.atEnd() {

					break batchLoop
				}
			}
		}
	} else {
		for {
			batch := bp.getNextBatch()
			go func(batch []Job, waiter *sync.WaitGroup) {
				bp.processBatch(batch, waiter)
			}(batch, &asyncProcessWait)
			if bp.atEnd() {
				break
			}
		}
	}
	return processListener
}

func (bp *BatchProcessor) BeginImmediate() {
	results := make(chan JobResult, len(bp.jobCache))
	go func() {
		for _, job := range bp.jobCache {
			result, err := job.Execute()
			results <- JobResult{Job: job, Result: result, Error: err, Success: err == nil}
		}
		results <- JobResult{Job: &endJob{}, Result: nil, Error: nil, Success: true}
	}()
	// shadow
	for result := range results {
		job := result.Job
		_, isEndJob := job.(*endJob)
		if isEndJob {
			break
		} else {
			callback := *bp.callback
			callback([]JobResult{result})
		}
	}
}
