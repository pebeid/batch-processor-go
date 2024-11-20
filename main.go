package main

import (
	"github.com/pebeid/batch-processor-go/batchjob"
	"github.com/pebeid/batch-processor-go/makejob"
)

func main() {
	processor, _ := batchjob.InstantiateBatchProcessor(10)
	printHelloJob := makejob.WithNoArgsAndNoReturn(printHello)
	processor.AddJob(printHelloJob)
	// executer := batchjob.BatchJob{}
	// executer.Execute(printHello)

}

func printHello() {
	println("HELLO WORLD!")
}
