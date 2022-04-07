package openApi

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type OpenApiConverter interface {
	GetData(Options) <-chan RawData
	AssignJobs(in <-chan RawData) <-chan Output
}

type Options struct {
	DBOptions
}

type TTLOpenApiConverter struct {
	Db
	WorkerCount  int //should be = number of cpus
	JobPerWorker int //max work each worker may do
}

func (c *TTLOpenApiConverter) GetData(opts Options) <-chan RawData {
	in := make(chan RawData, c.JobPerWorker)
	Debug("running getData with opts %v, maxJobPerWorker: %v", opts.DBOptions, c.JobPerWorker)
	c.Db.Read(in, opts.DBOptions)
	return in
}

func (c *TTLOpenApiConverter) AssignJobs(in <-chan RawData) <-chan Output {
	outs := []<-chan Output{}
	for i := 0; i < c.WorkerCount; i++ {
		outs = append(outs, c.work(in))
	}
	return c.merge(outs...)
}

func (c *TTLOpenApiConverter) merge(outs ...<-chan Output) <-chan Output {
	merged := make(chan Output, 1)
	var wg sync.WaitGroup
	wg.Add(len(outs))
	for _, ch := range outs {
		go func(ch <-chan Output) {
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

func (c *TTLOpenApiConverter) work(in <-chan RawData) <-chan Output {
	out := make(chan Output, 1)
	go func() {
		defer close(out)
		for d := range in {
			out <- c.convert(d)
		}
	}()
	return out
}

/**
`
{
  "openapi": "3.0.0",
  "info": {
    "title": "KIS API Specification",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://52.74.51.47/rest",
      "description": "KIS API Server"
    }
  ],
  "paths": {
	"uri":{
		"method": {
			"tags": TAG,
			"security": SECURITY,
			"parameters": PARAMETERS,
			"requestBody": REQUEST_BODY,
			"responses": RESPONSES
		}
	}
  }
}
`
*/

type Output struct {
	Uri     string
	Method  string
	Content string
}

func (c *TTLOpenApiConverter) convert(data RawData) Output {
	if m, err := regexp.MatchString("^(get:|post:|put:|delete:)/", data.uriPattern); !m || err != nil {
		Debug("invalid uri pattern format: %v %v", data.uriPattern, err)
		return Output{}
	}
	body := ""
	if data.parameters != "" && data.parameters != "[]" {
		body += fmt.Sprintf(`"parameters":%v,`, data.parameters)
	}
	if data.requestBody != "" && data.requestBody != "[]" && data.requestBody != "{}" {
		body += fmt.Sprintf(`"requestBody":%v,`, data.requestBody)
	}
	if data.security != "" && data.security != "[]" {
		body += fmt.Sprintf(`"security":[{"Bearer": []}],`)
	}
	data.tags = strings.ReplaceAll(data.tags, "Mas-rest-bridge", "Ttl-based")
	body += fmt.Sprintf(`"summary":%v, "tags":%v, "responses": %v`, strings.TrimPrefix(data.summary, "MAS_"), data.tags, data.responses)
	uri_method := strings.Split(data.uriPattern, ":")
	body = fmt.Sprintf(`{%v}`, body)
	return Output{Uri: uri_method[1], Method: uri_method[0], Content: body}
}
