package diary

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/utils"
	"time"
)

type Service struct {
	Repository RepositoryInterface
	Tasks      chan utils.Task
}

func NewService(repository RepositoryInterface, tasksChannel chan utils.Task) ServiceInterface {
	return &Service{Repository: repository, Tasks: tasksChannel}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) (err error) {
	return s.Repository.GetDbContext().ExecTx(fn)
}

func (service *Service) CreateDiary(diaryDto DiaryDto) (diaryDtoResult DiaryDto, err error) {
	err = service.withTransaction(func(tx *sql.Tx) (err error) {

		diaryDto.StartTime = time.Now()
		diaryDto.EndTime = utils.Optional[time.Time]{HasValue: false}

		currentDiaryPagination, err := service.Repository.GetDiary(DiaryFilter{}, 1, 1)
		if err != nil {
			return
		}

		if len(currentDiaryPagination.Items) > 0 {
			currentDiaryPagination.Items[0].EndTime = sql.NullTime{
				Time:  diaryDto.StartTime,
				Valid: true,
			}
			_, err = service.Repository.UpdateDiary(tx, currentDiaryPagination.Items[0])
			if err != nil {
				return
			}
		}

		diaryModel, err := diaryDto.ToModel()
		if err != nil {
			return
		}

		result, err := service.Repository.CreateDiary(tx, diaryModel)
		if err != nil {
			return
		}

		diaryDtoResult, err = result.ToDto()
		return
	})

	return

}

func (service *Service) GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryDto], error) {

	diaryModel, err := service.Repository.GetDiary(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[DiaryDto]{}, err
	}

	paginationReponse, err := ParsePaginationToDto(&diaryModel)

	if err != nil {
		return utils.PaginationResponse[DiaryDto]{}, err
	}

	return paginationReponse, nil
}

func (service *Service) UpdateDiary(diaryDto DiaryDto) (result bool, err error) {
	err = service.withTransaction(func(tx *sql.Tx) (err error) {
		diaryModel, err := diaryDto.ToModel()
		if err != nil {
			return
		}

		result, err = service.Repository.UpdateDiary(tx, diaryModel)

		return
	})

	return
}

func (service *Service) GetSummary() (DiarySummary, error) {
	filter := DiaryFilter{
		DateRange: utils.Optional[DateRange]{
			HasValue: true,
			Value: DateRange{
				Start: time.Now().Add(time.Hour * -1),
				End:   time.Now(),
			},
		},
	}

	diaryModelPagination, err := service.Repository.GetDiary(filter, 1, 200)

	if err != nil {
		return DiarySummary{}, err
	}

	diaryDtoPagination, err := ParsePaginationToDto(&diaryModelPagination)

	if err != nil {
		return DiarySummary{}, err
	}

	totalActivities := len(diaryDtoPagination.Items)
	TotalTimeSpentSeconds, err := calculateDailyDuration(diaryDtoPagination.Items)
	if err != nil {
		return DiarySummary{}, err
	}

	longestActivity, err := getLongestActivity(diaryDtoPagination.Items)
	if err != nil {
		return DiarySummary{}, err
	}

	return DiarySummary{
		Date:                  time.Now(),
		TotalActivities:       totalActivities,
		TotalTimeSpentSeconds: TotalTimeSpentSeconds,
		LongestActivity:       &longestActivity,
	}, nil
}

func calculateDailyDuration(diaryDtos []DiaryDto) (int, error) {
	totalDuration := 0

	for _, diaryDto := range diaryDtos {
		if diaryDto.EndTime.HasValue {
			duration := int(diaryDto.EndTime.Value.Sub(diaryDto.StartTime).Seconds())
			totalDuration += duration
		} else {
			duration := int(time.Since(diaryDto.StartTime).Seconds())
			totalDuration += duration
		}
	}

	return totalDuration, nil
}

func getLongestActivity(diaryDtos []DiaryDto) (LongestActivity, error) {
	longestActivity := LongestActivity{
		Name:              "",
		DurationSeconds:   0,
		DurationFormatted: "",
	}

	for _, diaryDto := range diaryDtos {
		if diaryDto.EndTime.HasValue {
			duration := int(diaryDto.EndTime.Value.Sub(diaryDto.StartTime).Seconds())
			if duration > longestActivity.DurationSeconds {
				longestActivity.DurationSeconds = duration
				longestActivity.Name = diaryDto.Name
			}
		} else {
			duration := int(time.Since(diaryDto.StartTime).Seconds())
			if duration > longestActivity.DurationSeconds {
				longestActivity.DurationSeconds = duration
				longestActivity.Name = diaryDto.Name
			}
		}
	}

	return longestActivity, nil
}

func (service *Service) DuplicateDiary(id int) (DiaryDto, error) {
	fmt.Println("Duplicating diary with ID:", id)
	filter := DiaryFilter{
		ID: utils.Optional[int]{
			HasValue: true,
			Value:    id,
		},
	}

	diaryModelPagination, err := service.Repository.GetDiary(filter, 1, 1)

	if err != nil {
		return DiaryDto{}, err
	}

	if len(diaryModelPagination.Items) == 0 {
		return DiaryDto{}, errors.New("diary not found")
	}

	diaryDtoCurrent, err := diaryModelPagination.Items[0].ToDto()

	if err != nil {
		return DiaryDto{}, err
	}

	diaryDtoNew, err := service.CreateDiary(diaryDtoCurrent)
	if err != nil {
		return DiaryDto{}, err
	}

	return diaryDtoNew, nil
}
