package main

import (
	"github.com/pebeid/batch-processor-go/batchjob"
)

func main() {
	executer := batchjob.BatchJob{}
	executer.Execute(printHello)
}

func printHello() {
	println("HELLO WORLD!")
}
