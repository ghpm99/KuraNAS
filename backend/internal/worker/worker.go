package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"

	"time"
)

type WorkerContext struct {
	Tasks      chan utils.Task
	Service    *files.Service
	Repository *files.Repository
}

func StartWorkers(context *WorkerContext, numWorkers int) {

	for i := range numWorkers {
		go worker(i, context)
	}

	go startWorkersScheduler(context)
}

func startWorkersScheduler(context *WorkerContext) {
	for {
		time.Sleep(1 * time.Hour) // ⏳ Roda a cada 10 minutos
		fmt.Println("Escaneamento de arquivos")
		context.Tasks <- utils.Task{
			Type: utils.ScanFiles,
			Data: "Escaneamento de arquivos",
		}
		fmt.Println("📁 Tarefa de escaneamento de arquivos enviada para a fila")
	}
}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		fmt.Printf("Worker %d: Processando tarefa %s\n", id, task.Data)

		if task.Type == utils.ScanFiles {
			ScanFilesHandler(context.Service)
		}
		fmt.Printf("Worker %d: Tarefa %s completa\n", id, task.Data)
	}
}
