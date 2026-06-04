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
	"nas-go/api/pkg/i18n"
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
	executors[StepTypeOllamaPull] = func(step jobs.StepModel) error {
		return executeOllamaPullStep(context, step)
	}
	executors[StepTypeAIPlaylistCluster] = func(step jobs.StepModel) error {
		return executeAIPlaylistClusterStep(context, step)
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

	metadata, err := getMetadata(fileDto, pythonScriptRunner, aiServiceForImageClassification(context))
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

	// Walk the tree and ask the DB about one file at a time. The database is
	// indexed by path and exists to be queried — so we do a small lookup per
	// file instead of loading the entire home_file table into a map in memory
	// (which does not scale past a few tens of thousands of files). A file is
	// enqueued only when it is genuinely new or its size/mtime changed.
	//
	// enqueued counts files that actually entered the processing pipeline as a
	// result of THIS scan. CreateJob returns id 0 (no error) when idempotency
	// skips a file that already has a pending job — those are not counted, so
	// the completion notification reflects the real number of files sent to the
	// pipeline rather than every candidate seen.
	enqueued := 0
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

		stat, exists, statErr := context.FilesService.GetFileStatByPath(path)
		if statErr != nil {
			return fmt.Errorf("lookup file stat for %q: %w", path, statErr)
		}

		if exists {
			sameSize := stat.Size == info.Size()
			// home_file.updated_at is TIMESTAMPTZ (microsecond precision in Postgres),
			// while filesystem ModTime has nanosecond precision. Comparing the raw
			// values always mismatches after a DB round-trip, flagging every file as
			// changed on every scan and re-enqueuing the whole library indefinitely.
			// Truncate both sides to second precision before comparing.
			sameModTime := !stat.UpdatedAt.IsZero() &&
				stat.UpdatedAt.Truncate(time.Second).Equal(info.ModTime().Truncate(time.Second))
			if sameSize && sameModTime {
				return nil
			}
		}

		fileDto := files.FileDto{
			Path:       path,
			ParentPath: filepath.Dir(path),
		}
		if parseErr := fileDto.ParseFileInfoToFileDto(info); parseErr != nil {
			return nil
		}

		plan, planErr := buildFileProcessingPlan(fileDto, JobTypeFSEvent, JobPriorityLow)
		if planErr != nil {
			log.Printf("[diff] skipping file %q: %v\n", path, planErr)
			return nil
		}

		jobID, createErr := context.JobOrchestrator.CreateJob(plan)
		if createErr != nil {
			return createErr
		}
		if jobID > 0 {
			enqueued++
		}

		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	if enqueued == 0 {
		return ErrStepSkipped
	}

	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_FILE_SCAN_COMPLETED_TITLE"),
		i18n.Translate("NOTIFICATION_FILE_SCAN_COMPLETED_MESSAGE", enqueued),
		"",
	)

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
