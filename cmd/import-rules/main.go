package main

import (
	"log"
	"os"
	"robot/internal/app"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./cmd/import-rules <excel-file>")
	}

	if err := app.LoadRulesFromExcel(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
