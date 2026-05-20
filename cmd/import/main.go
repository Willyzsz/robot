package main

import (
	"log"
	"os"
	"robot/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/import <excel-file>")
	}

	if err := app.LoadFromExcel(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}






