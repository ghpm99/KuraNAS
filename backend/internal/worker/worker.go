package worker

import (
	"fmt"
	"time"
)

type Task struct {
	Data string
}

var Tasks = make(chan Task, 100)

func StartWorkers(numWorkers int) {

	for i := range numWorkers {
		go worker(i, Tasks)
	}

	go startWorkersScheduler()
}

func startWorkersScheduler() {
	for {
		time.Sleep(10 * time.Minute) // ⏳ Roda a cada 10 minutos
		task := Task{Data: "Escaneamento de arquivos"}
		Tasks <- task
		fmt.Println("📁 Tarefa de escaneamento de arquivos enviada para a fila")
	}
}

func worker(id int, tasks <-chan Task) {
	for task := range tasks {
		fmt.Printf("Worker %d: Processando tarefa %s\n", id, task.Data)
		// Simula um trabalho demorado

		fmt.Printf("Worker %d: Tarefa %s completa\n", id, task.Data)
	}
}
