package openApi

import (
	"fmt"
	"regexp"
	"sync"
)

type OpenApiConverter interface {
	GetData(Options) (<-chan RawData, error)
	AssignJobs(in <-chan RawData) <-chan string
}

type Options struct {
	DBOptions
	JobPerWorker int //max work each worker may do
}

type TOpenApiConverter struct {
	Db
	WorkerCount int
}

func (c *TOpenApiConverter) GetData(opts Options) (<-chan RawData, error) {
	in := make(chan RawData, opts.JobPerWorker)
	err := c.Db.Read(in, opts.DBOptions)
	return in, err
}

func (c *TOpenApiConverter) AssignJobs(in <-chan RawData) <-chan string {
	outs := []<-chan string{}
	for i := 0; i < c.WorkerCount; i++ {
		outs = append(outs, c.work(in))
	}
	return c.merge(outs...)
}


func (c *TOpenApiConverter) merge(outs ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(outs))
	for _, ch := range outs {
		go func(ch <-chan string) {
			defer wg.Done()
			for v := range ch {
				merged <- v
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

func (c *TOpenApiConverter) work(in <-chan RawData) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer close(out)
		for d := range in {
			out <- c.convert(d)
		}
	}()
	return out
}

func (c *TOpenApiConverter) convert(data RawData) string {
	if m, err := regexp.MatchString("^(get:|post:|put:|delete:)", data.uriPattern); !m || err != nil {
		Debug("invalid uri pattern format: %v %v", data.uriPattern, err)
		return ""
	}
	return fmt.Sprint(data)
}
