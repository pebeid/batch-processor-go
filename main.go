package main

import (
	"github.com/pebeid/batch-processor-go/batchjob"
	"github.com/pebeid/batch-processor-go/batchjob/makejob"
)

func main() {
	processor, _ := batchjob.InstantiateBatchProcessor(10)
	printHelloJob := makejob.WithNoArgsAndNoReturn(printHello)
	printHelloJob2 := makejob.WithNoArgsAndNoReturn(printHello)
	processor.AddJob(printHelloJob)
	processor.AddJob(printHelloJob2)
	println(processor.Count())
	processor.RemoveJob(printHelloJob)
	println(processor.Count())
	// executer := batchjob.BatchJob{}
	// executer.Execute(printHello)

}

func printHello() {
	println("HELLO WORLD!")
}
