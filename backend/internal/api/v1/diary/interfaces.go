package diary

import (
	"database/sql"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *sql.DB
	CreateDiary(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error)
	GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error)
	UpdateDiary(transaction *sql.Tx, diary DiaryModel) (bool, error)
}

type ServiceInterface interface {
	CreateDiary(diaryDto DiaryDto) (diaryDtoResult DiaryDto, err error)
	GetFiles(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryDto], error)
}
