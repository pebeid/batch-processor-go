package main

import (
	"strconv"

	"github.com/pebeid/batch-processor-go/batchjob"
	"github.com/pebeid/batch-processor-go/batchjob/makejob"
)

func main() {
	processor, _ := batchjob.InstantiateBatchProcessor(10)
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
	processor.ExecuteImmediate(func(res batchjob.JobResult) {
		switch res.Success {
		case true:
			println("Count: " + strconv.Itoa(res.Result.(int)))
		case false:
			println("Error executing job:", res.Error)
		}
	})

}

// func printHello() {
// 	println("HELLO WORLD!")
// }

var count = 0

func getTimeStamp() (interface{}, error) {
	count++
	return count, nil
}
