package batchjob

import (
	"strconv"
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
		bp.Begin()
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

func (bp *BatchProcessor) Begin() {

	// empty job cache
	if len(bp.jobCache) == 0 {
		return
	}

	// Init result collector
	results := make(chan []JobResult)
	defer close(results)

	// Init batches
	var splitStart, splitEnd = bp.beginFrom, int(bp.batchSize)
	if bp.batchSize < 1 {
		splitEnd = len(bp.jobCache)
	}

	processBatch := func() (done bool) {
		if splitEnd > len(bp.jobCache) {
			splitEnd = len(bp.jobCache)
		}
		batch := bp.jobCache[splitStart:splitEnd]
		bp.batchesInProcess = bp.batchesInProcess + 1
		go func(batch []Job) {
			res := []JobResult{}
			for _, job := range batch {
				result, err := job.Execute()
				res = append(res, JobResult{Job: job, Result: result, Error: err, Success: err == nil})
			}
			results <- res
		}(batch)
		if splitEnd == len(bp.jobCache) {
			done = true
		} else {
			splitStart = splitEnd
			splitEnd = splitEnd + int(bp.batchSize)
			done = false
		}
		return done
	}

	bp.state = Active

	if bp.batchInterval != 0 {
		var done = false
	batchLoop:
		for {
			var alarm = <-time.Tick(bp.batchInterval)
			switch alarm {
			default:
				done = processBatch()
				if done {
					break batchLoop
				}
			}
		}
	} else {
		var done bool = false
		for {
			done = processBatch()
			if done {
				break
			}
		}
	}

	for result := range results {
		println("CALLING CALLBACK " + strconv.Itoa(len(result)))
		callback := *bp.callback
		callback(result)
		bp.batchesInProcess = bp.batchesInProcess - 1
		if bp.batchesInProcess <= 0 {
			bp.beginFrom = splitEnd
			return
		}
	}
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
