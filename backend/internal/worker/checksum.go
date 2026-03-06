package worker

import (
	"encoding/json"
	"fmt"
	"log"
	jobs "nas-go/api/internal/api/v1/jobs"
)

func UpdateCheckSumWorker(context *WorkerContext, data any) {
	if context == nil {
		log.Println("UpdateCheckSumWorker: worker context nulo")
		return
	}

	fileId, ok := data.(int)

	if !ok {
		log.Println("Erro ao converter ID do arquivo: data não é int")
		return
	}
	if fileId <= 0 {
		log.Println("UpdateCheckSumWorker: fileId invalido")
		return
	}

	if context.JobOrchestrator != nil {
		if _, err := context.JobOrchestrator.CreateJob(PlannedJob{
			Type:     JobTypeFSEvent,
			Priority: JobPriorityNormal,
			Scope: JobScope{
				FileID: &fileId,
			},
			Steps: []PlannedStep{
				{
					Key:         "checksum",
					Type:        StepTypeChecksum,
					MaxAttempts: 1,
					Payload:     mustMarshalChecksumStepPayload(fileId),
				},
			},
		}); err != nil {
			log.Printf("UpdateCheckSumWorker: erro ao criar job de checksum: %v\n", err)
		}
		return
	}

	// Legacy fallback for tests/contexts that still do not initialize
	// the orchestrator. It still uses the official checksum step executor.
	if context.FilesService == nil {
		log.Println("UpdateCheckSumWorker: files service nulo")
		return
	}

	payload, err := json.Marshal(StepFilePayload{FileID: fileId})
	if err != nil {
		log.Printf("UpdateCheckSumWorker: erro ao serializar payload de checksum: %v\n", err)
		return
	}

	if err := executeChecksumStep(context, jobs.StepModel{
		Type:    string(StepTypeChecksum),
		Payload: payload,
	}); err != nil {
		log.Printf("UpdateCheckSumWorker: erro no step de checksum: %v\n", err)
	}
}

func mustMarshalChecksumStepPayload(fileID int) []byte {
	payload, err := json.Marshal(StepFilePayload{FileID: fileID})
	if err != nil {
		log.Printf("UpdateCheckSumWorker: erro ao serializar payload (fallback vazio): %v\n", err)
		return []byte(fmt.Sprintf(`{"file_id":%d}`, fileID))
	}
	return payload
}
