package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/techx-mo/dbToSwagger/openApi"
)

func main() {
	flag.Parse()
	//write output to file
	{
		var data string = openApi.Do()
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
		if _, err := f.WriteString(data); err != nil {
			openApi.Debug("error write data to sql file: %v", err)
		}
	}
}
