package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/techx-mo/dbToSwagger/openApi"
)

func main() {
	flag.Parse()
	conn, err := sql.Open("mysql", os.Getenv("DB_PATH"))
	if err != nil {
		openApi.Log("error connecting db: %v", err)
		return
	}
	openApi.Log("Successfully connected mysql")
	defer conn.Close()
	var db openApi.Db = &openApi.OpenApiDb{Conn: conn}
	var converter openApi.OpenApiConverter = &openApi.TTLOpenApiConverter{
		Db:           db,
		WorkerCount:  8,
		JobPerWorker: 1,
	}
	data := converter.GetData(openApi.Options{
		DBOptions: openApi.DBOptions{ClientId: "kis-wts", Domain: "kis"},
	})
	out := converter.AssignJobs(data)
	// if len(out) == 0 {
	// 	openApi.Log("Got no records")
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
			  "Bearer": {
			    "type": "apiKey",
			    "in": "header",
			    "name": "authorization",
			  }
			}
		             },
		"paths": {%v}
	}
	`, strings.Join(swaggerPaths, ",")) //slice out last ,

	//write output to file
	{
		var outPath = os.Getenv("SWAGGER_OUT_PATH")
		if outPath == "" {
			outPath, _ = filepath.Abs("out.txt")
		}
		if err := os.Truncate(outPath, 0); err != nil {
			openApi.Debug("Failed to truncate: %v", err)
		}

		f, err := os.OpenFile(outPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			openApi.Log("error tructcate text file %v", err)
		}
		defer f.Close()
		if _, err := f.WriteString(swaggerFull); err != nil {
			openApi.Debug("error write data to sql file: %v", err)
		}
	}
}
