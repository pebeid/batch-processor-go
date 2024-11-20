package batchjob

type BatchJob struct {
}

func (bj *BatchJob) Execute(task func()) {
	task()
}
