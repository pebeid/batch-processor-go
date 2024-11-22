## Overview

`BatchProcessor` allows you to run jobs in batches in parallel.
You can also accumilate jobs and executes on the interval.

## Usage
First, you just define your jobs by implementing the `Job` interface method
```go
func Execute() (interface{}, error)
```

Once you have the `Job` implementation, you can create as many jobs and add them to the `BatchProcessor`.
Firstly, you need to instantiate a job using `batchjob.InstantiateBatchProcessor(...)` factory method

```go
batchProcessor, waitgroup, ok := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Interval: 2 * time.Second, Callback: &callback})
```

The options for setting up a batch process are:
* `BatchSize`: maximum size of each batch (note: batches will run even if have not reach maximum)
* `Interval`: `time.Duration` amount of time wait before collecting a batch and running it
* `Callback`: function with signature `func([]JobResult)` to process the results from running the jobs

`Callback` is mandatory, as well as either / or `BatchSize` and `Interval`.

The initialising function returns the batch processor object, a waitgroup and an ok flag to indicate
whether the batch processor was created successfully. _If the ok flag is false, the initiaiser is in
the wrong state, that is either `Callback` is missing or neither `BatchSize` nor `Interval`
is defined._

Once jobs are added via `AddJob(Job)` method call, the execution can being by calling `Begin()` method to run the jobs on the specified interval (default to 50ms), or `BeginImmediate()` method to override the Interval wait and batch and execute the Jobs as soons as possible.

> The `waitgroup.Wait()` function ([see documentation](https://pkg.go.dev/sync#WaitGroup)) must be called after the batch processor begins in order to prevent the system from shutting down before all batch processor go routines are executed. If the batch processor is used multiple times, the `Wait()` call must made after `Begin()` / `BeginImmediate` call.

Once a batch of jobs run, then the `Callback` function will execute with passed-in slice of results.
The `JobResult` struct holds the following information:
* `Job`: original job
* `Result`: any result returnd from the `Job.Execute()` function
* `Error`: any error returnd from the `Job.Execute()` function
* `Success`: `false` if `Job.Execute()` returned and error, otherwise `true`

**Note: If Jobs are added after `Begin()` call, they are automatically batched and executed on interval.**

## Complete Example
```go
// JOB
type Adder struct {
	terms []int
}

func (a *Adder) Execute() (interface{}, error) {
	var sum int = 0
	for _, term := range a.terms {
		sum = sum + term
	}
	return sum, nil
}

func SetAdder(numbers ...int) Adder {
	return Adder{terms: numbers}
}

// BatchProcessor setup

callback := func(results []batchjob.JobResult) {
    var accumulator int
    for _, result := range results {
        if result.Success {
            accumulator = accumulator + result.Result.(int)
        }
    }
    println("Batch sum: " + strconv.Itoa(accumulator))
}

processor, waiter, ok := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Interval: 2 * time.Second, Callback: &callback})

//Batch execution

if ok {
    processor.addJob(SetAdder(1, 2, 3))
    processor.addJob(SetAdder(10, 20, 30))
    processor.addJob(SetAdder(100, 200, 300))
    processor.Begin()
    waiter.Wait()
}
```