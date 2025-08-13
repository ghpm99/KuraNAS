package worker

import (
	"log"
	"sync"
)

func StartResultMonitorWorker(monitorChannel <-chan ResultWorkerData, wg *sync.WaitGroup) {
	defer wg.Done()

	totalProcessed := 0
	totalSuccess := 0
	totalErrors := 0

	for result := range monitorChannel {
		// Aqui você pode monitorar, logar ou tomar ações conforme o resultado
		totalProcessed++
		if result.Success {
			totalSuccess++
		} else {
			totalErrors++
			log.Printf("Arquivo: %s Error: %s", result.Path, result.Error)
		}
		log.Printf("Processados:%d Com sucesso:%d Com error:%d", totalProcessed, totalSuccess, totalErrors)
	}
}
