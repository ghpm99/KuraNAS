package worker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/utils"
)

type StepFilePayload struct {
	FileID int            `json:"file_id,omitempty"`
	Path   string         `json:"path,omitempty"`
	File   *files.FileDto `json:"file,omitempty"`
}

func buildStepExecutors(context *WorkerContext) map[StepType]StepExecutor {
	executors := map[StepType]StepExecutor{}

	executors[StepTypeMetadata] = func(step jobs.StepModel) error {
		return executeMetadataStep(context, step)
	}
	executors[StepTypeScanFilesystem] = func(step jobs.StepModel) error {
		return executeScanFilesystemStep(context, step)
	}
	executors[StepTypeDiffAgainstDB] = func(step jobs.StepModel) error {
		return executeDiffAgainstDBStep(context, step)
	}
	executors[StepTypeChecksum] = func(step jobs.StepModel) error {
		return executeChecksumStep(context, step)
	}
	executors[StepTypePersist] = func(step jobs.StepModel) error {
		return executePersistStep(context, step)
	}
	executors[StepTypeThumbnail] = func(step jobs.StepModel) error {
		return executeThumbnailStep(context, step)
	}
	executors[StepTypePlaylistIndex] = func(step jobs.StepModel) error {
		return executePlaylistIndexStep(context, step)
	}
	executors[StepTypeMarkDeleted] = func(step jobs.StepModel) error {
		return executeMarkDeletedStep(context, step)
	}
	executors[StepTypeTakeoutExtract] = func(step jobs.StepModel) error {
		return executeTakeoutExtractStep(context, step)
	}

	return executors
}

func decodeStepPayload(payloadRaw []byte) (StepFilePayload, error) {
	if len(payloadRaw) == 0 {
		return StepFilePayload{}, nil
	}

	payload := StepFilePayload{}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return StepFilePayload{}, fmt.Errorf("decode step payload: %w", err)
	}

	return payload, nil
}

func resolveFileDtoForStep(service files.ServiceInterface, payload StepFilePayload) (files.FileDto, error) {
	if payload.File != nil {
		return *payload.File, nil
	}

	if payload.FileID > 0 {
		fileDto, err := service.GetFileById(payload.FileID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return files.FileDto{}, ErrStepSkipped
			}
			return files.FileDto{}, err
		}
		return fileDto, nil
	}

	if payload.Path == "" {
		return files.FileDto{}, ErrStepSkipped
	}

	info, err := os.Stat(payload.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return files.FileDto{}, ErrStepSkipped
		}
		return files.FileDto{}, err
	}

	fileDto := files.FileDto{
		Path:       payload.Path,
		ParentPath: filepath.Dir(payload.Path),
	}
	if parseErr := fileDto.ParseFileInfoToFileDto(info); parseErr != nil {
		return files.FileDto{}, parseErr
	}

	persistedFile, persistedErr := service.GetFileByNameAndPath(fileDto.Name, fileDto.Path)
	if persistedErr == nil {
		return persistedFile, nil
	}
	if !errors.Is(persistedErr, sql.ErrNoRows) {
		return files.FileDto{}, persistedErr
	}

	return fileDto, nil
}

func executeMetadataStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("files service is required for metadata step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}

	fileDto, err := resolveFileDtoForStep(context.FilesService, payload)
	if err != nil {
		return err
	}

	metadata, err := getMetadata(fileDto, pythonScriptRunner, context.AIService)
	if err != nil {
		return err
	}
	if metadata == nil {
		return ErrStepSkipped
	}

	// If there is no persisted file ID yet, metadata extraction is still valid,
	// but persistence will happen in a dedicated step.
	if fileDto.ID <= 0 {
		log.Printf("[metadata] metadata extracted but file not persisted yet, skipping update (path=%s)\n", fileDto.Path)
		return nil
	}

	fileDto.Metadata = metadata
	updated, err := context.FilesService.UpdateFile(fileDto)
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("metadata step did not update file id=%d", fileDto.ID)
	}

	return nil
}

func executeChecksumStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("files service is required for checksum step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}

	fileDto, err := resolveFileDtoForStep(context.FilesService, payload)
	if err != nil {
		return err
	}

	checksum, err := getCheckSum(fileDto, utils.GetFileChecksum, utils.GetDirectoryChecksum)
	if err != nil {
		return err
	}

	if fileDto.ID <= 0 {
		return nil
	}

	fileDto.CheckSum = checksum
	updated, err := context.FilesService.UpdateFile(fileDto)
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("checksum step did not update file id=%d", fileDto.ID)
	}

	return nil
}

func executePersistStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("files service is required for persist step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}

	fileDto, err := resolveFileDtoForStep(context.FilesService, payload)
	if err != nil {
		return err
	}

	if fileDto.Name == "" || fileDto.Path == "" {
		return ErrStepSkipped
	}

	existingRecord, err := context.FilesService.GetFileByNameAndPath(fileDto.Name, fileDto.Path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, createErr := createFileRecord(context.FilesService, fileDto)
			return createErr
		}
		return err
	}

	_, err = UpdateFileRecord(context.FilesService, fileDto, existingRecord)
	return err
}

func executeThumbnailStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("files service is required for thumbnail step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}

	fileDto, err := resolveFileDtoForStep(context.FilesService, payload)
	if err != nil {
		return err
	}
	if fileDto.ID <= 0 {
		return ErrStepSkipped
	}

	CreateThumbnailWorker(context.FilesService, fileDto.ID, context.Logger)
	return nil
}

func executePlaylistIndexStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	if context == nil || context.VideoService == nil {
		return fmt.Errorf("video service is required for playlist index step")
	}

	GenerateVideoPlaylistsWorker(context.VideoService, context.Logger)
	return nil
}

func executeScanFilesystemStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil {
		return fmt.Errorf("worker context is required for scan step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}
	root := payload.Path
	if root == "" {
		return ErrStepSkipped
	}

	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrStepSkipped
		}
		return err
	}
	if !info.IsDir() {
		return ErrStepSkipped
	}

	return nil
}

func executeDiffAgainstDBStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil || context.JobOrchestrator == nil {
		return fmt.Errorf("files service and orchestrator are required for diff step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}
	root := payload.Path
	if root == "" {
		return ErrStepSkipped
	}

	// Batch-load all known files under root to avoid N+1 queries (Fix 5)
	knownFiles := map[string]files.FileDto{}
	filter := files.FileFilter{
		PathPrefix: utils.Optional[string]{HasValue: true, Value: root},
	}
	page := 1
	for {
		result, listErr := context.FilesService.GetFiles(filter, page, 500)
		if listErr != nil {
			return fmt.Errorf("batch load files under %q: %w", root, listErr)
		}
		for _, f := range result.Items {
			knownFiles[f.Path] = f
		}
		if !result.Pagination.HasNext {
			break
		}
		page++
	}

	changedPaths := []string{}
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			log.Printf("[diff] skipping inaccessible path %q: %v\n", path, walkErr)
			return nil
		}
		if d.IsDir() {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		existing, exists := knownFiles[path]
		if !exists {
			changedPaths = append(changedPaths, path)
			return nil
		}

		sameSize := existing.Size == info.Size()
		sameModTime := !existing.UpdatedAt.IsZero() && existing.UpdatedAt.UnixNano() == info.ModTime().UnixNano()
		if !sameSize || !sameModTime {
			changedPaths = append(changedPaths, path)
		}

		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	if len(changedPaths) == 0 {
		return ErrStepSkipped
	}

	for _, changedPath := range changedPaths {
		fileInfo, statErr := os.Stat(changedPath)
		if statErr != nil {
			continue
		}

		fileDto := files.FileDto{
			Path:       changedPath,
			ParentPath: filepath.Dir(changedPath),
		}
		if parseErr := fileDto.ParseFileInfoToFileDto(fileInfo); parseErr != nil {
			continue
		}

		plan, planErr := buildFileProcessingPlan(fileDto, JobTypeFSEvent, JobPriorityLow)
		if planErr != nil {
			log.Printf("[diff] skipping file %q: %v\n", changedPath, planErr)
			continue
		}

		_, createErr := context.JobOrchestrator.CreateJob(plan)
		if createErr != nil {
			return createErr
		}
	}

	return nil
}

func executeMarkDeletedStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("files service is required for mark_deleted step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}
	root := payload.Path
	if root == "" {
		return ErrStepSkipped
	}

	page := 1
	pageSize := 500
	updatedAny := false
	filter := files.FileFilter{
		PathPrefix: utils.Optional[string]{HasValue: true, Value: root},
	}
	for {
		result, listErr := context.FilesService.GetFiles(filter, page, pageSize)
		if listErr != nil {
			return listErr
		}

		for _, file := range result.Items {
			_, statErr := os.Stat(file.Path)
			missing := statErr != nil && errors.Is(statErr, os.ErrNotExist)

			if missing && !file.DeletedAt.HasValue {
				file.DeletedAt = utils.Optional[time.Time]{
					HasValue: true,
					Value:    time.Now(),
				}
				updated, updateErr := context.FilesService.UpdateFile(file)
				if updateErr != nil {
					return fmt.Errorf("mark missing file deleted id=%d: %w", file.ID, updateErr)
				}
				if !updated {
					return fmt.Errorf("mark missing file deleted id=%d: no rows updated", file.ID)
				}
				updatedAny = true
			}

			if !missing && file.DeletedAt.HasValue {
				file.DeletedAt = utils.Optional[time.Time]{HasValue: false}
				updated, updateErr := context.FilesService.UpdateFile(file)
				if updateErr != nil {
					return fmt.Errorf("restore file from deleted state id=%d: %w", file.ID, updateErr)
				}
				if !updated {
					return fmt.Errorf("restore file from deleted state id=%d: no rows updated", file.ID)
				}
				updatedAny = true
			}
		}

		if !result.Pagination.HasNext {
			break
		}
		page++
	}

	if !updatedAny {
		return ErrStepSkipped
	}

	return nil
}

func buildFileProcessingPlan(fileDto files.FileDto, jobType JobType, priority JobPriority) (PlannedJob, error) {
	persistPayload, err := marshalPayload(StepFilePayload{
		Path: fileDto.Path,
		File: &fileDto,
	})
	if err != nil {
		return PlannedJob{}, fmt.Errorf("marshal persist payload: %w", err)
	}
	commonPayload, err := marshalPayload(StepFilePayload{
		Path: fileDto.Path,
	})
	if err != nil {
		return PlannedJob{}, fmt.Errorf("marshal common payload: %w", err)
	}

	steps := []PlannedStep{
		{
			Key:         "persist",
			Type:        StepTypePersist,
			MaxAttempts: 3,
			Payload:     persistPayload,
		},
		{
			Key:         "metadata",
			Type:        StepTypeMetadata,
			DependsOn:   []string{"persist"},
			MaxAttempts: 3,
			Payload:     commonPayload,
		},
		{
			Key:         "checksum",
			Type:        StepTypeChecksum,
			DependsOn:   []string{"persist"},
			MaxAttempts: 3,
			Payload:     commonPayload,
		},
	}

	formatType := utils.GetFormatTypeByExtension(fileDto.Format)
	if formatType.Type == utils.FormatTypeImage || formatType.Type == utils.FormatTypeVideo {
		steps = append(steps, PlannedStep{
			Key:         "thumbnail",
			Type:        StepTypeThumbnail,
			DependsOn:   []string{"persist"},
			MaxAttempts: 3,
			Payload:     commonPayload,
		})
	}
	if formatType.Type == utils.FormatTypeVideo {
		steps = append(steps, PlannedStep{
			Key:         "playlist_index",
			Type:        StepTypePlaylistIndex,
			DependsOn:   []string{"persist"},
			MaxAttempts: 3,
			Payload:     commonPayload,
		})
	}

	return PlannedJob{
		Type:     jobType,
		Priority: priority,
		Scope: JobScope{
			Path: fileDto.Path,
		},
		Steps: steps,
	}, nil
}

func marshalPayload(v any) ([]byte, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshalPayload: %w", err)
	}
	return payload, nil
}
