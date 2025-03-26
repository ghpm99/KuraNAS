package main

import (
	"nas-go/api/internal/app"
	"nas-go/api/internal/worker"
)

func main() {
	application, err := app.InitializeApp()
	if err != nil {
		panic(err)
	}

	worker.StartWorkers(2)

	application.Run(":8080", false)
}
