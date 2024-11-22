package test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pebeid/batch-processor-go/batchjob"
	"github.com/pebeid/batch-processor-go/batchjob/makejob"
)

func matchInSlice(match []string, slice []string) bool {
	results := []int{}
	for _, m := range match {
		found := false
	sliceLoop:
		for si, s := range slice {
			if m == s {
				for _, r := range results {
					if r == si {
						continue sliceLoop
					}
				}
				found = true
				results = append(results, si)
				break
			}
		}
		if !found {
			return false
		}
	}
	return len(results) == len(match)
}

func Test_ExecuteImmediate(t *testing.T) {
	count := 0
	counter := func() {
		count++
	}

	batches := []string{}
	callback := func(res []batchjob.JobResult) {
		batches = append(batches, strconv.Itoa(len(res)))
	}

	processor, waiter, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Callback: &callback})
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.BeginImmediate()
	waiter.Wait()

	if count != 3 {
		t.Error("Expected 3 got " + strconv.Itoa(count))
	}

	if !matchInSlice([]string{"1", "2"}, batches) {
		t.Error("Did not get correct batching " + strings.Join(batches, ","))
	}
}

func Test_ExecuteBatches(t *testing.T) {

	count := 0
	counter := func() {
		count++
	}

	batches := []string{}
	callback := func(res []batchjob.JobResult) {
		batches = append(batches, strconv.Itoa(len(res)))
	}

	processor, waiter, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Callback: &callback})
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.Begin()
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	waiter.Wait()

	if count != 4 {
		t.Error("Expected 4 got " + strconv.Itoa(count))
	}

	if !matchInSlice([]string{"2", "2"}, batches) {
		t.Error("Did not get correct batching " + strings.Join(batches, ","))
	}
}

func Test_ExecuteAtInterval(t *testing.T) {

	count := 0
	counter := func() {
		count++
	}

	batches := []string{}
	callback := func(res []batchjob.JobResult) {
		batches = append(batches, strconv.Itoa(len(res)))
	}

	processor, waiter, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{Interval: 2 * time.Second, Callback: &callback})
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.Begin()
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	waiter.Wait()

	if count != 4 {
		t.Error("Expected 4 got " + strconv.Itoa(count))
	}
}

func Test_ExecuteBatchesAtInterval(t *testing.T) {

	count := 0
	counter := func() {
		count++
	}

	batches := []string{}
	callback := func(res []batchjob.JobResult) {
		batches = append(batches, strconv.Itoa(len(res)))
	}

	processor, waiter, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Interval: 2 * time.Second, Callback: &callback})
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.Begin()
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	waiter.Wait()

	if count != 4 {
		t.Error("Expected 4 got " + strconv.Itoa(count))
	}

	if !matchInSlice([]string{"2", "2"}, batches) {
		t.Error("Did not get correct batching " + strings.Join(batches, ","))
	}
}

func Test_ExecuteBatchesAtIntervalWithDelay(t *testing.T) {

	count := 0
	counter := func() {
		count++
	}

	batches := []string{}
	callback := func(res []batchjob.JobResult) {
		batches = append(batches, strconv.Itoa(len(res)))
	}

	processor, waiter, _ := batchjob.InstantiateBatchProcessor(batchjob.BatchProccessInitialiser{BatchSize: 2, Interval: 2 * time.Second, Callback: &callback})
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	processor.Begin()
	time.Sleep(5 * time.Second) // wait for 2 seconds before adding jobs to simulate delay
	processor.AddJob(makejob.WithNoArgsAndNoReturn(counter))
	waiter.Wait()

	if count != 4 {
		t.Error("Expected 4 got " + strconv.Itoa(count))
	}

	if !matchInSlice([]string{"2", "1", "1"}, batches) {
		t.Error("Did not get correct batching " + strings.Join(batches, ","))
	}
}
