package search

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/search"

	"github.com/lib/pq"
)

func (r *Repository) IsFuzzySearchAvailable() bool {
	r.fuzzySupportProbe.Do(func() {
		probeErr := r.scanRows(queries.CheckFuzzySearchSupportQuery, func(rows *sql.Rows) error {
			return rows.Scan(&r.isFuzzySupported)
		})
		if probeErr != nil {
			r.isFuzzySupported = false
		}
	})
	return r.isFuzzySupported
}

func (r *Repository) SearchFilesFuzzy(query string, limit int) ([]FileResultModel, error) {
	foldedQuery := normalizeFuzzyQuery(query)
	if foldedQuery == "" {
		return []FileResultModel{}, nil
	}
	results := []FileResultModel{}
	err := r.scanRows(queries.SearchFilesFuzzyQuery, func(rows *sql.Rows) error {
		var item FileResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Format, &item.Starred, &item.Size, &item.UpdatedAt, &item.IsCold); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, foldedQuery, limit, pq.Array(mediaFormats))
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar arquivos por similaridade: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchFoldersFuzzy(query string, limit int) ([]FolderResultModel, error) {
	foldedQuery := normalizeFuzzyQuery(query)
	if foldedQuery == "" {
		return []FolderResultModel{}, nil
	}
	results := []FolderResultModel{}
	err := r.scanRows(queries.SearchFoldersFuzzyQuery, func(rows *sql.Rows) error {
		var item FolderResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Starred, &item.Size, &item.UpdatedAt, &item.IsCold); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, foldedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar pastas por similaridade: %w", err)
	}
	return results, nil
}
