package worker

import (
	"encoding/json"
	"log"
	jobs "nas-go/api/internal/api/v1/jobs"
)

func UpdateCheckSumWorker(context *WorkerContext, data any) {
	if context == nil {
		log.Println("UpdateCheckSumWorker: worker context is nil")
		return
	}

	fileId, ok := data.(int)
	if !ok {
		log.Println("UpdateCheckSumWorker: data is not int")
		return
	}
	if fileId <= 0 {
		log.Println("UpdateCheckSumWorker: invalid fileId")
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
			log.Printf("UpdateCheckSumWorker: failed to create checksum job: %v\n", err)
		}
		return
	}

	// Legacy fallback for tests/contexts that still do not initialize
	// the orchestrator. It still uses the official checksum step executor.
	if context.FilesService == nil {
		log.Println("UpdateCheckSumWorker: files service is nil")
		return
	}

	payload, err := json.Marshal(StepFilePayload{FileID: fileId})
	if err != nil {
		log.Printf("UpdateCheckSumWorker: failed to serialize checksum payload: %v\n", err)
		return
	}

	if err := executeChecksumStep(context, jobs.StepModel{
		Type:    string(StepTypeChecksum),
		Payload: payload,
	}); err != nil {
		log.Printf("UpdateCheckSumWorker: checksum step error: %v\n", err)
	}
}

func mustMarshalChecksumStepPayload(fileID int) []byte {
	return mustMarshalPayload(StepFilePayload{FileID: fileID})
}
