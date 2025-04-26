package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"

	"time"
)

type WorkerContext struct {
	Tasks      chan utils.Task
	Service    *files.Service
	Repository *files.Repository
}

func StartWorkers(context *WorkerContext, numWorkers int) {
	if !config.AppConfig.EnableWorkers {
		return
	}
	for i := range numWorkers {
		go worker(i, context)
	}

	go startWorkersScheduler(context)
}

func startWorkersScheduler(context *WorkerContext) {
	for {
		fmt.Println("Escaneamento de arquivos")
		context.Tasks <- utils.Task{
			Type: utils.ScanFiles,
			Data: "Escaneamento de arquivos",
		}
		fmt.Println("📁 Tarefa de escaneamento de arquivos enviada para a fila")
		time.Sleep(12 * time.Hour) // ⏳ Roda a cada 10 horas
	}
}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		fmt.Printf("Worker %d: Processando tarefa %s\n", id, task.Data)

		if task.Type == utils.ScanFiles {
			ScanFilesWorker(context.Service)
		} else if task.Type == utils.ScanDir {
			ScanDirWorker(context.Service, task.Data)
		}
		fmt.Printf("Worker %d: Tarefa %s completa\n", id, task.Data)
	}
}
