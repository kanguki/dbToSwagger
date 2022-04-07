package openApi

import (
	"testing"
)

var maxJobPerWorker int = 10

func BenchmarkGetData(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping TestRead in short mode")
	}
	for n := 0; n < b.N; n++ {
		data := c.GetData(Options{
			DBOptions: DBOptions{ClientId: "kis-wts", Domain: "kis"},
		})
		_ = data
		// for v := range data {
		// 	b.Log(v)
		// }
	}

}

func TestGetData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRead in short mode")
	}
	data := c.GetData(Options{
		DBOptions: DBOptions{ClientId: "kis-wts", Domain: "kis"},
	})
	_ = data
	// for v := range data {
	// 	t.Log(v)
	// }

}

func BenchmarkAssignJobs(b *testing.B) {
	element := []RawData{{
		name: "get:/api/123",
	},
		{
			name: "/shouldFail",
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

	for n := 0; n < b.N; n++ {
		result := c.AssignJobs(fakeData)
		_ = result
		// for v := range result {
		// 	Log(v)
		// }
	}

}
