package diary

import (
	"context"
	"database/sql"
	"nas-go/api/pkg/utils"
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
