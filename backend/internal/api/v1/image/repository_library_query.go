package image

import (
	"fmt"
	"strconv"
	"strings"

	queries "nas-go/api/pkg/database/queries/image"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type libraryQueryBuilder struct {
	arguments []any
	clauses   []string
}

func newLibraryQueryBuilder() *libraryQueryBuilder {
	return &libraryQueryBuilder{arguments: []any{pq.Array(utils.ImageFormats)}}
}

func (builder *libraryQueryBuilder) bind(fragment string, values ...any) string {
	for position, value := range values {
		builder.arguments = append(builder.arguments, value)
		fragment = strings.ReplaceAll(fragment, "@"+strconv.Itoa(position+1), "$"+strconv.Itoa(len(builder.arguments)))
	}
	return strings.TrimSpace(fragment)
}

func (builder *libraryQueryBuilder) addClause(fragment string, values ...any) {
	builder.clauses = append(builder.clauses, builder.bind(fragment, values...))
}

func (builder *libraryQueryBuilder) addTextFilter(filter LibraryFilter) {
	hasName := filter.NameQuery != ""
	hasContent := filter.ContentQuery != ""
	switch {
	case hasName && hasContent && !filter.MustMatchNameAndContent:
		builder.addClause(queries.LibraryFilterNameOrContentQuery, likeContainsPattern(filter.NameQuery), likeContainsPattern(filter.ContentQuery))
	case hasName && hasContent:
		builder.addClause(queries.LibraryFilterNameQuery, likeContainsPattern(filter.NameQuery))
		builder.addClause(queries.LibraryFilterContentQuery, likeContainsPattern(filter.ContentQuery))
	case hasName:
		builder.addClause(queries.LibraryFilterNameQuery, likeContainsPattern(filter.NameQuery))
	case hasContent:
		builder.addClause(queries.LibraryFilterContentQuery, likeContainsPattern(filter.ContentQuery))
	}
}

func (builder *libraryQueryBuilder) addFilter(filter LibraryFilter) {
	builder.addTextFilter(filter)
	if len(filter.Categories) > 0 {
		builder.addClause(queries.LibraryFilterCategoryQuery, pq.Array(categoryNames(filter.Categories)))
	}
	if filter.OnlyStarred {
		builder.addClause(queries.LibraryFilterStarredQuery)
	}
	if len(filter.Formats) > 0 {
		builder.addClause(queries.LibraryFilterFormatQuery, pq.Array(filter.Formats))
	}
	if filter.TakenFrom != nil {
		builder.addClause(queries.LibraryFilterTakenFromQuery, *filter.TakenFrom)
	}
	if filter.TakenTo != nil {
		builder.addClause(queries.LibraryFilterTakenToQuery, *filter.TakenTo)
	}
	if filter.Camera != "" {
		builder.addClause(queries.LibraryFilterCameraQuery, filter.Camera)
	}
	if filter.Folder != "" {
		builder.addClause(queries.LibraryFilterFolderQuery, filter.Folder)
	}
	if filter.AlbumID > 0 {
		builder.addClause(queries.LibraryFilterAlbumQuery, filter.AlbumID)
	}
}

func (builder *libraryQueryBuilder) addPosition(query LibraryListQuery) {
	if query.TakenBefore != nil {
		builder.addClause(queries.LibrarySeekTakenBeforeQuery, *query.TakenBefore)
	}
	if query.NewerThan != nil {
		builder.addClause(queries.LibraryKeysetBeforeQuery, query.NewerThan.sortKey(), query.NewerThan.FileID)
	}
	if query.Cursor != nil {
		builder.addClause(queries.LibraryKeysetAfterQuery, query.Cursor.sortKey(), query.Cursor.FileID)
	}
}

func (builder *libraryQueryBuilder) assemble(selectFragment string, tailFragments ...string) string {
	var assembled strings.Builder
	assembled.WriteString(selectFragment)
	assembled.WriteString(queries.LibraryScopeQuery)
	for _, clause := range builder.clauses {
		assembled.WriteString("\n    AND ")
		assembled.WriteString(clause)
	}
	for _, tailFragment := range tailFragments {
		assembled.WriteString("\n")
		assembled.WriteString(tailFragment)
	}
	return assembled.String()
}

func buildLibraryListQuery(query LibraryListQuery) (string, []any, error) {
	orderFragment, err := libraryOrderFragment(query.Sort, query.Order)
	if err != nil {
		return "", nil, err
	}
	if query.NewerThan != nil && IsKeysetOrdering(query.Sort, query.Order) {
		orderFragment = queries.LibraryOrderTakenAtOldestFirstQuery
	}

	builder := newLibraryQueryBuilder()
	builder.addFilter(query.Filter)
	builder.addPosition(query)

	if IsKeysetOrdering(query.Sort, query.Order) {
		limitFragment := builder.bind(queries.LibraryLimitQuery, query.Limit)
		return builder.assemble(queries.LibraryListSelectQuery, orderFragment, limitFragment), builder.arguments, nil
	}

	limitFragment := builder.bind(queries.LibraryLimitOffsetQuery, query.Limit, query.Offset)
	return builder.assemble(queries.LibraryListSelectQuery, orderFragment, limitFragment), builder.arguments, nil
}

func buildLibraryCountQuery(filter LibraryFilter) (string, []any) {
	builder := newLibraryQueryBuilder()
	builder.addFilter(filter)
	return builder.assemble(queries.LibraryCountSelectQuery), builder.arguments
}

func buildLibraryTimelineQuery(filter LibraryFilter) (string, []any) {
	builder := newLibraryQueryBuilder()
	builder.addFilter(filter)
	builder.addClause(queries.LibraryFilterDatedOnlyQuery)
	return builder.assemble(queries.LibraryTimelineSelectQuery, queries.LibraryTimelineGroupQuery), builder.arguments
}

func libraryOrderFragment(sort LibrarySort, order LibrarySortOrder) (string, error) {
	isDescending := order == LibrarySortOrderDesc
	switch sort {
	case LibrarySortTakenAt:
		return pickByOrder(isDescending, queries.LibraryOrderTakenAtDescQuery, queries.LibraryOrderTakenAtAscQuery), nil
	case LibrarySortName:
		return pickByOrder(isDescending, queries.LibraryOrderNameDescQuery, queries.LibraryOrderNameAscQuery), nil
	case LibrarySortSize:
		return pickByOrder(isDescending, queries.LibraryOrderSizeDescQuery, queries.LibraryOrderSizeAscQuery), nil
	default:
		return "", fmt.Errorf("unsupported library sort: %q", sort)
	}
}

func pickByOrder(isDescending bool, descendingFragment string, ascendingFragment string) string {
	if isDescending {
		return descendingFragment
	}
	return ascendingFragment
}

func categoryNames(categories []ClassificationCategory) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, string(category))
	}
	return names
}

var likeWildcardEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func likeContainsPattern(text string) string {
	return "%" + likeWildcardEscaper.Replace(strings.ToLower(text)) + "%"
}
