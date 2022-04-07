package openApi

import (
	"database/sql"
	"flag"
	"testing"
)

var db Db
var c OpenApiConverter

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		var local = "admin:password@tcp(localhost:3306)/test"
		conn, err := sql.Open("mysql", local)
		if err != nil {
			Log("error connecting db: %v", err)
			return
		}
		Log("Successfully connected mysql")
		defer conn.Close()
		db = &OpenApiDb{Conn: conn}
		c = &TTLOpenApiConverter{Db: db, JobPerWorker: 1, WorkerCount: 2}
		//100 job workers consumes 10x 10 job workers wow
		//Bench results show lowest mem consumption when MaxJobPerWorker is 1 :D. 0 works similar as 1 too :D

		//WorkerCount = 2 results in very good performance :D

	}
	m.Run()
}
