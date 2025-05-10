package diary

import (
	"context"
	"database/sql"
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

func (s *Service) withTransaction(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := s.Repository.GetDbContext().BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = fn(tx); err != nil {
		return
	}

	return tx.Commit()
}

func (service *Service) CreateDiary(diaryDto DiaryDto) (diaryDtoResult DiaryDto, err error) {
	err = service.withTransaction(context.Background(), func(tx *sql.Tx) (err error) {
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
	err = service.withTransaction(context.Background(), func(tx *sql.Tx) (err error) {
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
	return DiarySummary{
		Date:                    time.Now(),
		TotalActivities:         4,
		TotalTimeSpentSeconds:   457,
		TotalTimeSpentFormatted: "teste",
		LongestActivity: &LongestActivity{
			Name:              "teste atividade",
			DurationSeconds:   400,
			DurationFormatted: "teste 2",
		},
	}, nil
}
