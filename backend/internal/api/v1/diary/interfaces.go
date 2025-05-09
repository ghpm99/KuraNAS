package diary

import "database/sql"

type RepositoryInterface interface {
	GetDbContext() *sql.DB
	CreateDiary(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error)
}

type ServiceInterface interface {
	CreateDiary(diaryDto DiaryDto) (diaryDtoResult DiaryDto, err error)
}
