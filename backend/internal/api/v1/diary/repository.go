package diary

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/diary"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
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

	query := queries.InsertDiaryQuery

	var diaryId int
	err := transaction.QueryRow(
		query,
		args...,
	).Scan(&diaryId)

	if err != nil {
		return fail(err)
	}

	diary.ID = diaryId

	return diary, nil
}

func (r *Repository) GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
	paginationResponse := utils.PaginationResponse[DiaryModel]{
		Items: []DiaryModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	// A lógica de construção dos argumentos pode ser mantida aqui
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
		!filter.DateRange.HasValue,
		filter.DateRange.Value.Start,
		filter.DateRange.Value.End,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta e escaneamento é movida para dentro desta função
		rows, err := tx.Query(
			queries.GetDiaryQuery,
			args...,
		)
		if err != nil {
			return err
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
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, diary)
		}

		return nil
	})

	if err != nil {
		// Retorna a resposta de paginação vazia e o erro em caso de falha na transação
		return paginationResponse, fmt.Errorf("falha ao obter diário: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
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
