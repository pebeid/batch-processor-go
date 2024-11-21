package main

import (
	"time"

	"github.com/pebeid/batch-processor-go/batchjob"
	"github.com/pebeid/batch-processor-go/batchjob/makejob"
)

func main() {
	callback := func(ress []batchjob.JobResult) {
		for _, res := range ress {
			switch res.Success {
			case true:
				println("Count: " + res.Result.(time.Time).String())
			case false:
				println("Error executing job:", res.Error)
			}
		}
	}

	processor, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{Interval: time.Duration(5 * time.Second), Callback: &callback})
	// printHelloJob := makejob.WithNoArgsAndNoReturn(printHello)
	// printHelloJob2 := makejob.WithNoArgsAndNoReturn(printHello)
	// processor.AddJob(printHelloJob)
	// processor.AddJob(printHelloJob2)
	// println(processor.Count())
	// processor.RemoveJob(printHelloJob)
	// println(processor.Count())
	// executer := batchjob.BatchJob{}
	// executer.Execute(printHello)
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))
	processor.Begin()
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))
	processor.AddJob(makejob.WithNoArgs(getTimeStamp))

}

// func printHello() {
// 	println("HELLO WORLD!")
// }

func getTimeStamp() (interface{}, error) {
	time.Tick(1 * time.Second)
	return time.Now(), nil
}
