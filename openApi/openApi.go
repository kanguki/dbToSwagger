package openApi

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

var TokenType = map[string]string{"jwt": "jwt", "otpToken": "otpToken"}

func Do() string {
	conn, err := sql.Open("mysql", os.Getenv("DB_PATH"))
	if err != nil {
		Log("error connecting db: %v", err)
		return ""
	}
	Log("Successfully connected mysql")
	defer conn.Close()
	var db Db = &OpenApiDb{Conn: conn}
	var converter OpenApiConverter = &TTLOpenApiConverter{
		Db:           db,
		WorkerCount:  8,
		JobPerWorker: 1,
	}
	clientId, domain := os.Getenv("CLIENT_ID"), os.Getenv("DOMAIN")
	data := converter.GetData(Options{
		DBOptions: DBOptions{ClientId: clientId, Domain: domain},
	})
	out := converter.AssignJobs(data)
	// if len(out) == 0 {
	// 	Log("Got no records")
	// 	return
	// }
	m := make(map[string]map[string]string) //[uri][[methods]]
	for v := range out {
		if v.Uri == "" || v.Method == "" { //got error
			continue
		}
		uri, method, content := v.Uri, v.Method, v.Content
		if m[uri] == nil {
			m[uri] = make(map[string]string)
		}
		m[uri][method] = content

	}
	swaggerPaths := []string{}
	for k, v := range m {
		methodContent := []string{}
		for method, content := range v {
			methodContent = append(methodContent, fmt.Sprintf(`"%v": %v`, method, content))
		}
		swaggerPaths = append(swaggerPaths, fmt.Sprintf(`
		"%v": {
			%v
		}
		`, k, strings.Join(methodContent, ",")))
	}
	swaggerFull := fmt.Sprintf(`
	{
		"openapi": "3.0.0",
		"info": {
		  "title": "KIS API Specification",
		  "version": "1.0.0"
		},
		"servers": [
		  {
		    "url": "https://beta.kisvn.vn:8443/rest",
		    "description": "KIS API Server"
		  }
		],
		"components": {
			"securitySchemes": {
			  "%v": {
			    "type": "apiKey",
			    "in": "header",
			    "name": "authorization",
			  },
			  "%v": {
			    "type": "apiKey",
			    "in": "header",
			    "name": "otpToken",
			  }
			}
		             },
		"paths": {%v}
	}
	`, TokenType["jwt"], TokenType["otpToken"], strings.Join(swaggerPaths, ",")) //slice out last ,
	return swaggerFull
}

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
		if strings.Contains(data.forward_data, `"tokenType": "VERIFIED"`) {
			body += fmt.Sprintf(`"security":[{"%v": []},{"%v": []}],`, TokenType["jwt"], TokenType["otpToken"])
		} else {
			body += fmt.Sprintf(`"security":[{"%v": []}],`, TokenType["jwt"])
		}
	}
	data.tags = strings.ReplaceAll(data.tags, "Mas-rest-bridge", "Ttl-based")
	body += fmt.Sprintf(`"summary":"%v", "tags":%v, "responses": %v`, strings.TrimPrefix(data.summary, "MAS_"), data.tags, data.responses)
	uri_method := strings.Split(data.uriPattern, ":")
	body = fmt.Sprintf(`{%v}`, body)
	return Output{Uri: uri_method[1], Method: uri_method[0], Content: body}
}
