package diary

import (
	"nas-go/api/pkg/utils"
	"time"
)

type DiaryDto struct {
	ID          int                       `json:"id"`
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	StartTime   time.Time                 `json:"start_time"`
	EndTime     utils.Optional[time.Time] `json:"end_time"`
	Duration    int                       `json:"duration"`
}

func (diaryDto *DiaryDto) CalculateDuration() {
	if diaryDto.EndTime.HasValue {
		diaryDto.Duration = int(diaryDto.EndTime.Value.Sub(diaryDto.StartTime).Seconds())
	} else {
		diaryDto.Duration = int(time.Since(diaryDto.StartTime).Seconds())
	}
}

type LongestActivity struct {
	Name              string `json:"name"`
	DurationSeconds   int    `json:"duration_seconds"`
	DurationFormatted string `json:"duration_formatted"`
}

type DiarySummary struct {
	Date                  time.Time        `json:"date"`
	TotalActivities       int              `json:"total_activities"`
	TotalTimeSpentSeconds int              `json:"total_time_spent_seconds"`
	LongestActivity       *LongestActivity `json:"longest_activity,omitempty"`
}

func (diaryModel *DiaryModel) ToDto() (DiaryDto, error) {
	diaryDto := DiaryDto{
		ID:          diaryModel.ID,
		Name:        diaryModel.Name,
		Description: diaryModel.Description,
		StartTime:   diaryModel.StartTime,
	}
	if err := diaryDto.EndTime.ParseFromNullTime(diaryModel.EndTime); err != nil {
		return diaryDto, err
	}

	return diaryDto, nil
}

type DateRange struct {
	Start time.Time
	End   time.Time
}

type DiaryFilter struct {
	ID          utils.Optional[int]
	Name        utils.Optional[string]
	Description utils.Optional[string]
	StartTime   utils.Optional[time.Time]
	EndTime     utils.Optional[time.Time]
	DateRange   utils.Optional[DateRange]
}

func ParsePaginationToDto(pagination *utils.PaginationResponse[DiaryModel]) (utils.PaginationResponse[DiaryDto], error) {
	paginationResponse := utils.PaginationResponse[DiaryDto]{
		Items: []DiaryDto{},
		Pagination: utils.Pagination{
			Page:     pagination.Pagination.Page,
			PageSize: pagination.Pagination.PageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	for _, fileModel := range pagination.Items {
		fileDtoResult, err := fileModel.ToDto()

		if err != nil {
			return paginationResponse, err
		}
		fileDtoResult.CalculateDuration()
		paginationResponse.Items = append(paginationResponse.Items, fileDtoResult)
	}
	paginationResponse.Pagination = pagination.Pagination

	return paginationResponse, nil
}
