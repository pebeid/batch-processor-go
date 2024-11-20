package batchjob

type Job interface {
	Execute()
}

type batchProcessorInner struct {
	jobCache   []Job
	batchCount int16
}

type BatchProcessor batchProcessorInner

func InstantiateBatchProcessor(batchCount int16) (BatchProcessor, bool) {
	return BatchProcessor{jobCache: []Job{}, batchCount: batchCount}, true
}

// type Job struct {
// }

// func (bj *Job) Execute(task func()) {
// 	task()
// }
