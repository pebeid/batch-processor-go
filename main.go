package main

import (
	"github.com/pebeid/batch-processor-go/batchjob"
)

func main() {
	processor, ok := batchjob.InstantiateBatchProcessor(10)
	// executer := batchjob.BatchJob{}
	// executer.Execute(printHello)
}

func printHello() {
	println("HELLO WORLD!")
}
