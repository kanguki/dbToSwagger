package openApi

import "testing"

var maxJobPerWorker int = 10
func TestAssignJobs(t *testing.T) {
	var c OpenApiConverter = &TOpenApiConverter{WorkerCount: 100}
	element := []RawData{{
		name:        "get:/api/123",
	},
	{
		name:        "/shouldFail",
	}}
	var makeFakeData = func(e []RawData) <-chan RawData {
		fakeData := make(chan RawData, maxJobPerWorker) 
		go func(e []RawData) {
			defer close(fakeData)
			for i := 0; i < maxJobPerWorker/len(e); i++ {
				for _, v := range e {
					fakeData <- v
				}
			}
		}(e)
		return fakeData
	}
	fakeData := makeFakeData(element)
	
	result := c.AssignJobs(fakeData)
	for v := range result {
		Log(v)
	}

}