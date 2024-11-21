package batchjob

import (
	"strconv"
	"sync"
	"time"
)

// JobResult
//
//		    Job 	original job passed
//	   	Result 	the result object of the job execution
//	   	Error 	the error encountered during the job execution
//	   	Success	true if the job execution was successful, false otherwise
//			This struct is used to pass the result of the job execution back to the caller.
type JobResult struct {
	Job     Job
	Result  interface{}
	Error   error
	Success bool
}

// Job
//
//	Execute() method is called when the BatchProcessor executes the job and returns a result as defined by the implementing functions as well as an error if there was one
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
	waiter           *sync.WaitGroup
	state            state
	beginFrom        int
	batchesInProcess uint
}

// BatchProcessorInitialiser
// Factory struct for holding the configuration of the batch job processor (BatchProcessor)
// Once configured. the struct is passed to BatchProcessor initialiser (i.e. InstantiateBatchProcessor)
// BatchProcessor can be configured to run jobs in batches based on a time interval (i.e. every x amount of time passed)
// or it can split all jobs into batches, with each batch run in a separate thread
// or if both BatchSize and Interval are specified, then the batch job will execite a batch of jobs, limited in size by
// BatchSize every Interval amount of time. Callback function is required to provess the batch running result.
//
//	BatchSize 	maximum number of jobs to run at once
//	Interval 	time between cycles of collecting next set of jobs into a batch (optionally) and running them
//	Callback	function that is able to handle the result of []Job run
type BatchProccessInitialiser struct {
	BatchSize uint16
	Interval  time.Duration
	Callback  *func([]JobResult)
}

// BatchProcessor
// Processes slices of Job(s) at a time by calling Job.Execute() on each job and returning the result and any error to the callback function
// Job(s) are run in multiple batches, each run in a goprocess
type BatchProcessor batchProcessorInner

// InstantiateBatchProcessor
// Returns BatchProcessor configured to run a maximum Job batch size and interval according to the initialiser object
//
//	Initialiser		BatchProcessor Callback function, BatchSize and Interval configuration
//
// Returns *BatchProcessor pointer, the *sync.WaitGroup pointer and bool to represent whether the BatchProcessor was initialised correctly
// If the bool is false, then the initialiser (BatchProcessorInitialiser) was not configured correctly.
// BatchProcessorInitialiser.Callback is mandatory and either BatchProcessorInitialiser.BatchSize AND/OR BatchProcessorInitialiser.Interval must
// be set
//
// The Wait() method on the sync.WaitGroup must be called by the Begin() method caller after. If it's not called, the system may exit before
// all the batches have had a chance to execute. If the BatchProcessor is used intermitently, the Wait must be called after each use.
func InstantiateBatchProcessor(initialiser BatchProccessInitialiser) (*BatchProcessor, *sync.WaitGroup, bool) {
	if initialiser.Callback == nil || (initialiser.BatchSize == 0 && initialiser.Interval == 0) {
		return nil, nil, false
	}
	var waiter sync.WaitGroup
	if initialiser.Interval == 0 {
		initialiser.Interval = 50 * time.Millisecond
	}
	return &BatchProcessor{jobCache: []Job{}, batchSize: initialiser.BatchSize, batchInterval: initialiser.Interval, callback: initialiser.Callback, waiter: &waiter}, &waiter, true
}

// AddJob
// Adds job to the queue, which is split into batches and run.
// If the BatchProcessor is already running (i.e. Begin() method is called at least once) then the added Job will execute in the
// next available batch
func (bp *BatchProcessor) AddJob(job Job) {
	bp.jobCache = append(bp.jobCache, job)
	if bp.state == Active && bp.batchesInProcess < 1 {
		println("Restarting Process")
		bp.Begin()
	}
}

// RemoveJob
// Removes Job from job queue. If a Job is added while the BatchProcessor is running, it may execute before it has the chance to be removed.
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

// Count
// Current Job queue length
func (bp *BatchProcessor) Count() int {
	return len(bp.jobCache)
}

// IsRunning
// Whether the BatchProcessor is Running
func (bp *BatchProcessor) IsRunning() bool {
	return bool(bp.state)
}

func (bp *BatchProcessor) getNextBatch() []Job {
	println("Getting next batch", strconv.Itoa(len(bp.jobCache)), bp.beginFrom)
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

func (bp *BatchProcessor) signalBatchStart() {
	bp.waiter.Add(1)
	bp.batchesInProcess = bp.batchesInProcess + 1
}

func (bp *BatchProcessor) signalBatchEnd() {
	bp.waiter.Done()
	bp.batchesInProcess = bp.batchesInProcess - 1
}

func (bp *BatchProcessor) processBatch(batch []Job) {
	println("Processing batch", strconv.Itoa(len(batch)))
	results := []JobResult{}
	for _, job := range batch {
		result, err := job.Execute()
		results = append(results, JobResult{Job: job, Result: result, Error: err, Success: err == nil})
	}
	callback := *bp.callback
	callback(results)
	bp.signalBatchEnd()
}

func (bp *BatchProcessor) atEnd() bool {
	return bp.beginFrom >= len(bp.jobCache)
}

// Begin
// Starts the interval countdown and the subsequent batching of Jobs and their running.
// IMPORTANT: Rembmer to call the Wait() method from the sync.WaitGroup return by the BatchProcessor initialiser.
func (bp *BatchProcessor) Begin() {
	println("BEGIN")

	bp.state = Active

	bp.waiter.Add(1)
	go func() {
	batchLoop:
		for {
			var alarm = <-time.Tick(bp.batchInterval)
			switch alarm {
			default:
				batch := bp.getNextBatch()
				bp.signalBatchStart()
				go func(batch []Job) {
					bp.processBatch(batch)
				}(batch)
				if bp.atEnd() {
					break batchLoop
				}
			}
		}
		bp.waiter.Done()
	}()
}

// BeginImmediate
// Executes the batches of Job without any interval
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
