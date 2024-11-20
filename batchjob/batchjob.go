package batchjob

// JobResult
type JobResult struct {
	result  interface{}
	err     error
	success bool
}

// Job interface
type Job interface {
	Execute() (interface{}, error)
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

func (bp *BatchProcessor) RemoveJob(job Job) {
}
