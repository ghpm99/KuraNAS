package documenttext

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/documenttext"

	"github.com/lib/pq"
)

const (
	contentWindowLeadingCharacters = 100
	contentWindowCharacters        = 400
	matchedRowsOpening             = "WITH matched AS ("
	matchedRowsClosing             = ")"
)

func buildSearchDocumentsQuery(query DocumentSearchQuery) (string, []any) {
	builder := &documentQueryBuilder{}
	columnsFragment := builder.bind(queries.SearchDocumentsColumnsQuery, pq.Array(query.TermPatterns.AllPatterns))
	builder.addClause(queries.SearchDocumentsFilterMatchQuery, query.TermPatterns.DrivingPattern, pq.Array(query.TermPatterns.AllPatterns))
	matchedRows := builder.assemble(columnsFragment, query.Filter, query.FolderPath, queries.SearchDocumentsOrderQuery)
	pageFragment := builder.bind(queries.DocumentsPageQuery, query.Limit, query.Offset)
	contentHitFragment := builder.bind(queries.SearchDocumentsContentHitQuery, pq.Array(query.Terms), contentWindowLeadingCharacters, contentWindowCharacters)
	statement := matchedRowsOpening + "\n" + matchedRows + "\n" + pageFragment + "\n" + matchedRowsClosing + "\n" + contentHitFragment
	return statement, builder.arguments
}

// SearchDocuments returns documents whose name or text contains every term,
// name matches first, then most recently updated. Only a bounded window of the
// stored text around the first match leaves the database.
func (repository *Repository) SearchDocuments(query DocumentSearchQuery) ([]DocumentMatchModel, error) {
	statement, arguments := buildSearchDocumentsQuery(query)
	matches := []DocumentMatchModel{}
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(statement, arguments...)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var match DocumentMatchModel
			var page sql.NullInt64
			if scanErr := rows.Scan(&match.FileID, &match.Name, &match.Path, &match.ParentPath, &match.Format, &match.Size, &match.UpdatedAt, &match.IsStarred, &match.HasContentMatch, &match.ContentWindow, &page); scanErr != nil {
				return scanErr
			}
			if page.Valid {
				pageNumber := int(page.Int64)
				match.Page = &pageNumber
			}
			matches = append(matches, match)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("documenttext: search documents: %w", err)
	}
	return matches, nil
}
