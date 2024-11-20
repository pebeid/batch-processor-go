package batchjob

// JobResult
type JobResult struct {
	result  interface{}
	err     error
	success bool
}

// Job interface
type Job interface {
	Execute() JobResult
}

type genericJob struct {
	function func(...interface{}) (interface{}, error)
	args     []interface{}
}

func (gj *genericJob) Execute() JobResult {
	res, err := gj.function(gj.args)
	if err != nil {
		return JobResult{result: res, success: false, err: err}
	} else {
		return JobResult{result: res, success: true}
	}
}

func MakeJob(function func(...interface{}) (result interface{}, err error), args ...interface{}) Job {
	return &genericJob{function: function, args: args}
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
