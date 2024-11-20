package makejob

import bj "github.com/pebeid/batch-processor-go/batchjob"

type genericJob struct {
	function func(...interface{}) (interface{}, error)
	args     []interface{}
}

func (gj *genericJob) Execute() (interface{}, error) {
	return gj.function(gj.args)
}

func WithArgs(function func(...interface{}) (result interface{}, err error), args ...interface{}) bj.Job {
	return &genericJob{function: function, args: args}
}

func WithNoArgs(function func() (result interface{}, err error)) bj.Job {
	modifiedFn := func(...interface{}) (result interface{}, err error) {
		return function()
	}
	return &genericJob{function: modifiedFn}
}

func WithArgsAndNoReturn(function func(...interface{}), args ...interface{}) bj.Job {
	modifiedFn := func(...interface{}) (result interface{}, err error) {
		function(args)
		return nil, nil
	}
	return &genericJob{function: modifiedFn}
}

func WithNoArgsAndNoReturn(function func() (result interface{}, err error)) bj.Job {
	modifiedFn := func(...interface{}) (result interface{}, err error) {
		function()
		return nil, nil
	}
	return &genericJob{function: modifiedFn}
}
