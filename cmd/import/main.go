package main

import (
	"log"
	"robot/internal/app"
)

func main() {
	if err := app.LoadFromExcel("form.xlsx"); err != nil {
		log.Fatal(err)
	}
}






