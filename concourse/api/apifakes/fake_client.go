// Code generated by counterfeiter. DO NOT EDIT.
package apifakes

import (
	"sync"

	"github.com/concourse/concourse-pipeline-resource/concourse/api"
)

type FakeClient struct {
	PipelinesStub        func(string) ([]api.Pipeline, error)
	pipelinesMutex       sync.RWMutex
	pipelinesArgsForCall []struct {
		arg1 string
	}
	pipelinesReturns struct {
		result1 []api.Pipeline
		result2 error
	}
	pipelinesReturnsOnCall map[int]struct {
		result1 []api.Pipeline
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) Pipelines(arg1 string) ([]api.Pipeline, error) {
	fake.pipelinesMutex.Lock()
	ret, specificReturn := fake.pipelinesReturnsOnCall[len(fake.pipelinesArgsForCall)]
	fake.pipelinesArgsForCall = append(fake.pipelinesArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("Pipelines", []interface{}{arg1})
	fake.pipelinesMutex.Unlock()
	if fake.PipelinesStub != nil {
		return fake.PipelinesStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.pipelinesReturns.result1, fake.pipelinesReturns.result2
}

func (fake *FakeClient) PipelinesCallCount() int {
	fake.pipelinesMutex.RLock()
	defer fake.pipelinesMutex.RUnlock()
	return len(fake.pipelinesArgsForCall)
}

func (fake *FakeClient) PipelinesArgsForCall(i int) string {
	fake.pipelinesMutex.RLock()
	defer fake.pipelinesMutex.RUnlock()
	return fake.pipelinesArgsForCall[i].arg1
}

func (fake *FakeClient) PipelinesReturns(result1 []api.Pipeline, result2 error) {
	fake.PipelinesStub = nil
	fake.pipelinesReturns = struct {
		result1 []api.Pipeline
		result2 error
	}{result1, result2}
}

func (fake *FakeClient) PipelinesReturnsOnCall(i int, result1 []api.Pipeline, result2 error) {
	fake.PipelinesStub = nil
	if fake.pipelinesReturnsOnCall == nil {
		fake.pipelinesReturnsOnCall = make(map[int]struct {
			result1 []api.Pipeline
			result2 error
		})
	}
	fake.pipelinesReturnsOnCall[i] = struct {
		result1 []api.Pipeline
		result2 error
	}{result1, result2}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.pipelinesMutex.RLock()
	defer fake.pipelinesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ api.Client = new(FakeClient)