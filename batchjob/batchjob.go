package batchjob

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
	jobCache  []Job
	batchSize int16
}

type BatchProcessor batchProcessorInner

func InstantiateBatchProcessor(batchSize int16) (BatchProcessor, bool) {
	return BatchProcessor{jobCache: []Job{}, batchSize: batchSize}, true
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

func (bp *BatchProcessor) ExecuteImmediate(callback func(JobResult)) {
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
