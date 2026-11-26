package analytics

import (
	"database/sql"
	"fmt"
	"time"

	"nas-go/api/pkg/database/queries/analytics"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
)

type ProcessingStage string

const (
	ProcessingStageChecksum     ProcessingStage = "checksum"
	ProcessingStageMetadata     ProcessingStage = "metadata"
	ProcessingStageThumbnail    ProcessingStage = "thumbnail"
	ProcessingStageDocumentText ProcessingStage = "document_text"
	ProcessingStageAI           ProcessingStage = "ai"
)

var processingFailureMessageKeys = map[ProcessingStage]string{
	ProcessingStageChecksum:     "ANALYTICS_PROCESSING_FAILURE_CHECKSUM",
	ProcessingStageMetadata:     "ANALYTICS_PROCESSING_FAILURE_METADATA",
	ProcessingStageThumbnail:    "ANALYTICS_PROCESSING_FAILURE_THUMBNAIL",
	ProcessingStageDocumentText: "ANALYTICS_PROCESSING_FAILURE_DOCUMENT_TEXT",
	ProcessingStageAI:           "ANALYTICS_PROCESSING_FAILURE_AI",
}

type ProcessingFailureModel struct {
	FileID   int
	Name     string
	Stage    ProcessingStage
	RawError string
	FailedAt time.Time
}

type ProcessingFailureDto struct {
	FileID   int       `json:"file_id"`
	Name     string    `json:"name"`
	Stage    string    `json:"stage"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}

func (r *Repository) GetProcessingFailuresPage(stage ProcessingStage, page int, pageSize int) (utils.PaginationResponse[ProcessingFailureModel], error) {
	failures := []ProcessingFailureModel{}
	var stageFilter any
	if stage != "" {
		stageFilter = string(stage)
	}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.ProcessingFailuresPageQuery,
			stageFilter,
			pageSize+1,
			utils.CalculateOffset(page, pageSize),
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var failure ProcessingFailureModel
			if err := rows.Scan(&failure.FileID, &failure.Name, &failure.Stage, &failure.RawError, &failure.FailedAt); err != nil {
				return err
			}
			failures = append(failures, failure)
		}
		return rows.Err()
	})
	if err != nil {
		return utils.PaginationResponse[ProcessingFailureModel]{}, fmt.Errorf("analytics processing failures page failed: %w", err)
	}
	return buildProcessingFailuresPage(failures, page, pageSize), nil
}

func buildProcessingFailuresPage(failures []ProcessingFailureModel, page int, pageSize int) utils.PaginationResponse[ProcessingFailureModel] {
	hasNext := len(failures) > pageSize
	if hasNext {
		failures = failures[:pageSize]
	}
	return utils.PaginationResponse[ProcessingFailureModel]{
		Items: failures,
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  hasNext,
			HasPrev:  page > 1,
		},
	}
}

func (s *Service) GetProcessingFailuresPage(stage ProcessingStage, page int, pageSize int) (utils.PaginationResponse[ProcessingFailureDto], error) {
	pageModels, err := s.Repository.GetProcessingFailuresPage(stage, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[ProcessingFailureDto]{}, err
	}
	failureDtos := make([]ProcessingFailureDto, 0, len(pageModels.Items))
	for _, failure := range pageModels.Items {
		failureDtos = append(failureDtos, toProcessingFailureDto(failure))
	}
	return utils.PaginationResponse[ProcessingFailureDto]{
		Items:      failureDtos,
		Pagination: pageModels.Pagination,
	}, nil
}

func toProcessingFailureDto(failure ProcessingFailureModel) ProcessingFailureDto {
	return ProcessingFailureDto{
		FileID:   failure.FileID,
		Name:     failure.Name,
		Stage:    string(failure.Stage),
		Error:    i18n.GetMessage(processingFailureMessageKeys[failure.Stage]),
		FailedAt: failure.FailedAt,
	}
}
