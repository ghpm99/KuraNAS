package files

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"
)

type RecentFileRepository struct {
	DbContext *database.DbContext
}

func NewRecentFileRepository(db *database.DbContext) *RecentFileRepository {
	return &RecentFileRepository{DbContext: db}
}

func (r *RecentFileRepository) Upsert(ip string, fileID int) error {
	// ExecTx gerencia o lock de escrita e a transação.
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.UpsertRecentFileQuery,
			ip, fileID,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("falha ao realizar upsert de arquivo recente: %w", err)
	}
	return nil
}

func (r *RecentFileRepository) DeleteOld(ip string, keep int) error {
	// ExecTx gerencia o lock de escrita e a transação.
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.DeleteOldRecentFilesQuery,
			ip, ip, keep,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("falha ao deletar arquivos recentes antigos: %w", err)
	}
	return nil
}

func (r *RecentFileRepository) GetRecentFiles(page int, pageSize int) ([]RecentFileModel, error) {
	var result []RecentFileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetRecentFilesQuery,
			pageSize+1,
			utils.CalculateOffset(page, pageSize),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rf RecentFileModel
			if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
				return err
			}
			result = append(result, rf)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao obter arquivos recentes: %w", err)
	}
	return result, nil
}

func (r *RecentFileRepository) Delete(ip string, fileID int) error {
	// ExecTx gerencia o lock de escrita e a transação.
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.DeleteRecentFileQuery,
			ip, fileID,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("falha ao deletar arquivo recente: %w", err)
	}
	return nil
}

func (r *RecentFileRepository) GetByFileID(fileID int) ([]RecentFileModel, error) {
	var result []RecentFileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetRecentByFileIDQuery,
			fileID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rf RecentFileModel
			if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
				return err
			}
			result = append(result, rf)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao obter arquivos recentes por ID do arquivo: %w", err)
	}
	return result, nil
}
