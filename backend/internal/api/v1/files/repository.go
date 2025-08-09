package files

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"
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

func (r *Repository) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {

	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
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
		!filter.Path.HasValue,
		filter.Path.Value,
		!filter.ParentPath.HasValue,
		filter.ParentPath.Value,
		!filter.Format.HasValue,
		filter.Format.Value,
		!filter.Type.HasValue,
		filter.Type.Value,
		!filter.DeletedAt.HasValue,
		filter.DeletedAt.Value,
		filter.Category,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta e escaneamento é movida para dentro desta função
		rows, err := tx.Query(
			queries.GetFilesQuery,
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			if err := rows.Scan(
				&file.ID,
				&file.Name,
				&file.Path,
				&file.ParentPath,
				&file.Format,
				&file.Size,
				&file.UpdatedAt,
				&file.CreatedAt,
				&file.LastInteraction,
				&file.LastBackup,
				&file.Type,
				&file.CheckSum,
				&file.DeletedAt,
				&file.Starred,
			); err != nil {
				return err
			}

			paginationResponse.Items = append(paginationResponse.Items, file)
		}

		return nil
	})

	if err != nil {
		// Retorna a resposta de paginação vazia e o erro em caso de falha na transação
		return paginationResponse, fmt.Errorf("falha na consulta de arquivos: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

func (r *Repository) CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error) {

	fail := func(err error) (FileModel, error) {
		return file, fmt.Errorf("CreateFile: %v", err)
	}

	args := []any{
		file.Name,
		file.Path,
		file.ParentPath,
		file.Format,
		file.Size,
		file.UpdatedAt,
		file.CreatedAt,
		file.LastInteraction,
		file.LastBackup,
		file.DeletedAt,
		file.Type,
		file.CheckSum,
	}

	query := queries.InsertFileQuery

	data, err := transaction.Exec(
		query,
		args...,
	)

	if err != nil {
		return fail(err)
	}

	fileId, err := data.LastInsertId()

	if err != nil {
		return fail(err)
	}

	file.ID = int(fileId)

	return file, nil
}

func (r *Repository) UpdateFile(transaction *sql.Tx, file FileModel) (bool, error) {
	fail := func(err error) (bool, error) {
		return false, fmt.Errorf("UpdateFile: %v", err)
	}

	result, err := transaction.Exec(
		queries.UpdateFileQuery,
		&file.Name,
		&file.Path,
		&file.ParentPath,
		&file.Format,
		&file.Size,
		&file.UpdatedAt,
		&file.CreatedAt,
		&file.LastInteraction,
		&file.LastBackup,
		&file.Type,
		&file.CheckSum,
		&file.DeletedAt,
		&file.Starred,
		&file.ID,
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

func (r *Repository) GetDirectoryContentCount(fileId int, parentPath string) (int, error) {
	var childrenCount int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {

		row := tx.QueryRow(
			queries.GetChildrenCountQuery,
			parentPath,
			fileId,
		)

		if err := row.Scan(&childrenCount); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("falha ao obter contagem de diretório: %w", err)
	}

	return childrenCount, nil
}

func (r *Repository) GetCountByType(fileType FileType) (int, error) {

	var count int
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {

		row := tx.QueryRow(
			queries.CountByTypeQuery,
			fileType,
		)

		if err := row.Scan(&count); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("GetCountByType: %v", err)
	}

	return count, nil
}

func (r *Repository) GetTotalSpaceUsed() (int, error) {
	var totalSpaceUsed int

	// Usa QueryTx para gerenciar o lock de leitura e a transação.
	// A lógica de consulta é movida para dentro desta função anônima.
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.TotalSpaceUsedQuery)

		// Escaneia o resultado
		if err := row.Scan(&totalSpaceUsed); err != nil {
			// Retorna o erro para que QueryTx o capture e o propague
			return err
		}

		// Se tudo correr bem, retorna nil para indicar sucesso
		return nil
	})

	if err != nil {
		// Trata o erro aqui. A função QueryTx já o formatou para nós.
		return 0, fmt.Errorf("falha ao obter espaço total usado: %w", err)
	}

	return totalSpaceUsed, nil
}

func (r *Repository) GetReportSizeByFormat() ([]SizeReportModel, error) {
	var report []SizeReportModel

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta é movida para dentro desta função anônima
		rows, err := tx.Query(queries.CountByFormatQuery, File)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Itera sobre os resultados da consulta
		for rows.Next() {
			var item SizeReportModel
			if err := rows.Scan(&item.Format, &item.Total, &item.Size); err != nil {
				return err
			}
			report = append(report, item)
		}

		// Se a iteração terminar sem erros, retorna nil
		return nil
	})

	if err != nil {
		// Retorna a slice vazia e o erro em caso de falha na transação
		return nil, fmt.Errorf("falha ao obter relatório por formato: %w", err)
	}

	return report, nil
}

func (r *Repository) GetTopFilesBySize(limit int) ([]FileModel, error) {
	var topFiles []FileModel

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta é movida para dentro desta função anônima
		rows, err := tx.Query(queries.TopFilesBySizeQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Itera sobre os resultados da consulta
		for rows.Next() {
			var file FileModel
			if err := rows.Scan(
				&file.ID,
				&file.Name,
				&file.Size,
				&file.Path,
			); err != nil {
				return err
			}
			topFiles = append(topFiles, file)
		}

		// Se a iteração terminar sem erros, retorna nil
		return nil
	})

	if err != nil {
		// Retorna a slice vazia e o erro em caso de falha na transação
		return nil, fmt.Errorf("falha ao obter top arquivos por tamanho: %w", err)
	}

	return topFiles, nil
}

func (r *Repository) GetDuplicateFiles(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {
	paginationResponse := utils.PaginationResponse[DuplicateFilesModel]{
		Items: []DuplicateFilesModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	// A lógica de construção dos argumentos pode ser mantida aqui
	args := []any{
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta e escaneamento é movida para dentro desta função
		rows, err := tx.Query(
			queries.GetDuplicateFilesQuery,
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var duplicate DuplicateFilesModel
			if err := rows.Scan(
				&duplicate.Name,
				&duplicate.Size,
				&duplicate.Copies,
				&duplicate.Paths,
			); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, duplicate)
		}

		return nil
	})

	if err != nil {
		// Retorna a resposta de paginação vazia e o erro em caso de falha na transação
		return paginationResponse, fmt.Errorf("falha ao obter arquivos duplicados: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}
