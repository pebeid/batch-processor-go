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
type batchProcessorInner struct {
	jobCache         []Job
	batchSize        uint16
	batchInterval    time.Duration
	batchesInProcess uint
}

type BatchProccessInitialiser struct {
	BatchSize uint16
	Interval  time.Duration
}

type BatchProcessor batchProcessorInner

func InstantiateBatchProcessor(initialiser BatchProccessInitialiser) (BatchProcessor, bool) {
	return BatchProcessor{jobCache: []Job{}, batchSize: initialiser.BatchSize, batchInterval: initialiser.Interval}, true
}

func (bp *BatchProcessor) AddJob(job Job) {
	bp.jobCache = append(bp.jobCache, job)
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

func (bp *BatchProcessor) alarmRunner(run func() bool) bool {
	var alarm = <-time.Tick(bp.batchInterval)
	switch alarm {
	default:
		return run()
	}
}

func (bp *BatchProcessor) Begin(callback func([]JobResult)) {
	results := make(chan []JobResult)
	defer close(results)
	var splitStart, splitEnd = 0, int(bp.batchSize)

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

	if bp.batchInterval != 0 {
		var done = false
		for {
			done = bp.alarmRunner(processBatch)
			if done {
				break
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

	// shadow
	for result := range results {
		println("CALLING CALLBACK " + strconv.Itoa(len(result)))
		callback(result)
		bp.batchesInProcess = bp.batchesInProcess - 1
		if bp.batchesInProcess <= 0 {
			return
		}
	}
}

func (bp *BatchProcessor) BeginImmediate(callback func(JobResult)) {
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
			callback(result)
		}
	}
}
