package diary

import (
	"database/sql"
	"errors"
	"fmt"
	queries "nas-go/api/pkg/database/queries/diary"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *sql.DB
}

func (repository *Repository) CreateDiary(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error) {
	fail := func(err error) (DiaryModel, error) {
		return diary, fmt.Errorf("CreateDiary: %v", err)
	}

	args := []any{
		diary.Name,
		diary.Description,
		diary.StartTime,
	}

	query := queries.CreateTableQuery

	data, err := transaction.Exec(
		query,
		args...,
	)

	if err != nil {
		return fail(err)
	}

	diaryId, err := data.LastInsertId()

	if err != nil {
		return fail(err)
	}

	diary.ID = int(diaryId)

	return diary, nil
}

func (repository *Repository) GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
	paginationReponse := utils.PaginationResponse[DiaryModel]{
		Items: []DiaryModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		!filter.ID.HasValue,
		filter.ID.Value,
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.Description.HasValue,
		filter.Description.Value,
		!filter.StartTime.HasValue,
		filter.StartTime.Value,
		!filter.EndTime.HasValue,
		filter.EndTime.Value,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	rows, err := repository.DbContext.Query(
		queries.GetDiaryQuery,
		args...,
	)

	if err != nil {
		return paginationReponse, err
	}

	defer rows.Close()

	for rows.Next() {
		var diary DiaryModel
		if err := rows.Scan(
			&diary.ID,
			&diary.Name,
			&diary.Description,
			&diary.StartTime,
			&diary.EndTime,
		); err != nil {
			return paginationReponse, err
		}

		paginationReponse.Items = append(paginationReponse.Items, diary)
	}

	paginationReponse.UpdatePagination()

	return paginationReponse, nil
}

func (repository *Repository) UpdateDiary(transaction *sql.Tx, diary DiaryModel) (bool, error) {
	fail := func(err error) (bool, error) {
		return false, fmt.Errorf("UpdateDiary: %v", err)
	}

	result, err := transaction.Exec(
		queries.UpdateDiaryQuery,
		&diary.Name,
		&diary.Description,
		&diary.StartTime,
		&diary.EndTime,
		&diary.ID,
	)

	if err != nil {
		return fail(err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fail(err)
	}

	if rowsAffected > 1 {
		transaction.Rollback()
		return fail(errors.New("multiple rows affected"))
	}

	return rowsAffected == 1, nil
}
