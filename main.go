package main

import (
	"database/sql"
	"flag"
	"os"

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
	var converter openApi.OpenApiConverter = &openApi.TOpenApiConverter{
		Db:          db,
		WorkerCount: 10,
	}
	data, err := converter.GetData(openApi.Options{
		DBOptions:    openApi.DBOptions{ClientId: "kis-wts", Domain: "kis"},
		JobPerWorker: 100,
	})
	if err != nil {
		openApi.Log(err.Error())
	}
	out := converter.AssignJobs(data)
	for v := range out {
		openApi.Log(v)
	}
	
}
