package image

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

const uniqueViolationCode = "23505"

type AlbumRepository struct {
	Db *database.DbContext
}

func NewAlbumRepository(db *database.DbContext) *AlbumRepository {
	return &AlbumRepository{Db: db}
}

type albumRowScanner interface {
	Scan(destinations ...any) error
}

func scanAlbum(scanner albumRowScanner) (AlbumModel, error) {
	var album AlbumModel
	var coverFileID sql.NullInt64
	err := scanner.Scan(&album.ID, &album.Name, &album.ItemCount, &coverFileID, &album.CreatedAt, &album.UpdatedAt)
	if err != nil {
		return AlbumModel{}, err
	}
	if coverFileID.Valid {
		coverID := int(coverFileID.Int64)
		album.CoverFileID = &coverID
	}
	return album, nil
}

func (r *AlbumRepository) ListAlbums(limit int, offset int) ([]AlbumModel, error) {
	albums := []AlbumModel{}
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.AlbumSelectQuery+"\n"+queries.AlbumOrderPageQuery, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			album, err := scanAlbum(rows)
			if err != nil {
				return err
			}
			albums = append(albums, album)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListAlbums: %w", err)
	}
	return albums, nil
}

func (r *AlbumRepository) GetAlbum(albumID int) (AlbumModel, error) {
	var album AlbumModel
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		album, scanErr = scanAlbum(tx.QueryRow(queries.AlbumSelectQuery+"\n"+queries.AlbumFilterIDQuery, albumID))
		return scanErr
	})
	if errors.Is(err, sql.ErrNoRows) {
		return AlbumModel{}, ErrAlbumNotFound
	}
	if err != nil {
		return AlbumModel{}, fmt.Errorf("GetAlbum: %w", err)
	}
	return album, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pq.Error
	return errors.As(err, &postgresError) && postgresError.Code == uniqueViolationCode
}

func (r *AlbumRepository) CreateAlbum(name string) (int, error) {
	var albumID int
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.AlbumInsertQuery, name).Scan(&albumID)
	})
	if isUniqueViolation(err) {
		return 0, ErrAlbumNameTaken
	}
	if err != nil {
		return 0, fmt.Errorf("CreateAlbum: %w", err)
	}
	return albumID, nil
}

func (r *AlbumRepository) UpdateAlbum(update AlbumUpdate) error {
	var newName, newCoverFileID any
	if update.Name != nil {
		newName = *update.Name
	}
	if update.CoverFileID != nil {
		newCoverFileID = *update.CoverFileID
	}

	var updatedRows int64
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		execution, err := tx.Exec(queries.AlbumUpdateQuery, update.AlbumID, newName, newCoverFileID)
		if err != nil {
			return err
		}
		updatedRows, err = execution.RowsAffected()
		return err
	})
	if isUniqueViolation(err) {
		return ErrAlbumNameTaken
	}
	if err != nil {
		return fmt.Errorf("UpdateAlbum: %w", err)
	}
	if updatedRows == 0 {
		return ErrAlbumCoverNotInAlbum
	}
	return nil
}

func (r *AlbumRepository) DeleteAlbum(albumID int) error {
	var deletedRows int64
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		execution, err := tx.Exec(queries.AlbumDeleteQuery, albumID)
		if err != nil {
			return err
		}
		deletedRows, err = execution.RowsAffected()
		return err
	})
	if err != nil {
		return fmt.Errorf("DeleteAlbum: %w", err)
	}
	if deletedRows == 0 {
		return ErrAlbumNotFound
	}
	return nil
}

func (r *AlbumRepository) AddAlbumItems(albumID int, fileIDs []int) (int, error) {
	return r.changeAlbumItems("AddAlbumItems", queries.AlbumItemsAddQuery, albumID, fileIDs, pq.Array(utils.ImageFormats))
}

func (r *AlbumRepository) RemoveAlbumItems(albumID int, fileIDs []int) (int, error) {
	return r.changeAlbumItems("RemoveAlbumItems", queries.AlbumItemsRemoveQuery, albumID, fileIDs)
}

func (r *AlbumRepository) changeAlbumItems(operation string, query string, albumID int, fileIDs []int, extraArguments ...any) (int, error) {
	arguments := append([]any{albumID, pq.Array(fileIDs)}, extraArguments...)

	var changedRows int64
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		execution, err := tx.Exec(query, arguments...)
		if err != nil {
			return err
		}
		changedRows, err = execution.RowsAffected()
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return int(changedRows), nil
}
