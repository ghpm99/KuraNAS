package documenttext

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	queries "nas-go/api/pkg/database/queries/documenttext"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type documentQueryBuilder struct {
	arguments []any
	clauses   []string
}

func (builder *documentQueryBuilder) bind(fragment string, values ...any) string {
	for position, value := range values {
		builder.arguments = append(builder.arguments, value)
		fragment = strings.ReplaceAll(fragment, "@"+strconv.Itoa(position+1), "$"+strconv.Itoa(len(builder.arguments)))
	}
	return strings.TrimSpace(fragment)
}

func (builder *documentQueryBuilder) addClause(fragment string, values ...any) {
	builder.clauses = append(builder.clauses, builder.bind(fragment, values...))
}

func (builder *documentQueryBuilder) addFilter(filter DocumentFilter, folderPath string) {
	if folderPath != "" {
		builder.addClause(queries.DocumentsFilterFolderQuery, folderPath)
	}
	if filter.ModifiedFrom != nil {
		builder.addClause(queries.DocumentsFilterModifiedFromQuery, *filter.ModifiedFrom)
	}
	if filter.ModifiedTo != nil {
		builder.addClause(queries.DocumentsFilterModifiedBeforeQuery, filter.ModifiedTo.AddDate(0, 0, 1))
	}
	if filter.FavoritesOnly {
		builder.addClause(queries.DocumentsFilterStarredQuery)
	}
	if filter.IndexState != "" {
		builder.addClause(queries.DocumentsFilterIndexStateQuery, filter.IndexState)
	}
}

func (builder *documentQueryBuilder) assemble(columnsFragment string, filter DocumentFilter, folderPath string, suffixFragments ...string) string {
	var assembled strings.Builder
	assembled.WriteString(columnsFragment)
	assembled.WriteString("\n")
	assembled.WriteString(builder.bind(queries.DocumentsBaseQuery, pq.Array(formatsOfGroup(filter.Format))))
	builder.addFilter(filter, folderPath)
	for _, clause := range builder.clauses {
		assembled.WriteString("\n    AND ")
		assembled.WriteString(clause)
	}
	for _, suffixFragment := range suffixFragments {
		assembled.WriteString("\n")
		assembled.WriteString(suffixFragment)
	}
	return assembled.String()
}

func formatsOfGroup(group string) []string {
	if group == "" {
		return DocumentFormats()
	}
	groupFormats := []string{}
	for _, format := range DocumentFormats() {
		if documentGroupByFormat[format] == group {
			groupFormats = append(groupFormats, format)
		}
	}
	return groupFormats
}

var orderFragmentBySort = map[string][2]string{
	SortName:       {queries.DocumentsOrderNameAscQuery, queries.DocumentsOrderNameDescQuery},
	SortSize:       {queries.DocumentsOrderSizeAscQuery, queries.DocumentsOrderSizeDescQuery},
	SortModifiedAt: {queries.DocumentsOrderModifiedAtAscQuery, queries.DocumentsOrderModifiedAtDescQuery},
	SortCreatedAt:  {queries.DocumentsOrderCreatedAtAscQuery, queries.DocumentsOrderCreatedAtDescQuery},
}

func orderFragmentOf(sort string, order string) (string, error) {
	fragments, isKnownSort := orderFragmentBySort[sort]
	if !isKnownSort {
		return "", fmt.Errorf("unsupported document sort: %q", sort)
	}
	if order == OrderDescending {
		return fragments[1], nil
	}
	return fragments[0], nil
}

func buildListDocumentsQuery(query DocumentListQuery) (string, []any, error) {
	orderFragment, err := orderFragmentOf(query.Sort, query.Order)
	if err != nil {
		return "", nil, err
	}
	builder := &documentQueryBuilder{}
	statement := builder.assemble(queries.ListDocumentsColumnsQuery, query.Filter, query.FolderPath, orderFragment)
	pageFragment := builder.bind(queries.DocumentsPageQuery, query.PageSize+1, utils.CalculateOffset(query.Page, query.PageSize))
	return statement + "\n" + pageFragment, builder.arguments, nil
}

func buildCountDocumentsQuery(filter DocumentFilter, folderPath string) (string, []any) {
	builder := &documentQueryBuilder{}
	statement := builder.assemble(queries.CountDocumentsColumnsQuery, filter, folderPath)
	return statement, builder.arguments
}

func (repository *Repository) ListDocuments(query DocumentListQuery) (utils.PaginationResponse[DocumentListItemModel], error) {
	response := utils.PaginationResponse[DocumentListItemModel]{
		Items:      []DocumentListItemModel{},
		Pagination: utils.Pagination{Page: query.Page, PageSize: query.PageSize},
	}
	statement, arguments, err := buildListDocumentsQuery(query)
	if err != nil {
		return response, fmt.Errorf("documenttext: list documents: %w", err)
	}

	err = repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(statement, arguments...)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var item DocumentListItemModel
			if scanErr := rows.Scan(&item.ID, &item.Name, &item.Format, &item.ParentPath, &item.Size, &item.ModifiedAt, &item.CreatedAt, &item.Starred, &item.IndexState); scanErr != nil {
				return scanErr
			}
			response.Items = append(response.Items, item)
		}
		return rows.Err()
	})
	if err != nil {
		return response, fmt.Errorf("documenttext: list documents: %w", err)
	}

	response.UpdatePagination()
	return response, nil
}

func (repository *Repository) CountDocuments(filter DocumentFilter, folderPath string) (int, error) {
	statement, arguments := buildCountDocumentsQuery(filter, folderPath)
	var total int
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(statement, arguments...).Scan(&total)
	})
	if err != nil {
		return 0, fmt.Errorf("documenttext: count documents: %w", err)
	}
	return total, nil
}

func (repository *Repository) FindActiveFolderPath(folderID int) (string, bool, error) {
	var folderPath string
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.SelectActiveFolderPathQuery, folderID).Scan(&folderPath)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("documenttext: find folder: %w", err)
	}
	return folderPath, true, nil
}
