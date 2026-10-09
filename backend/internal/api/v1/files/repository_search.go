package files

import (
	"fmt"
	"strconv"
	"strings"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type searchQueryBuilder struct {
	arguments []any
	clauses   []string
}

func (builder *searchQueryBuilder) bind(fragment string, values ...any) string {
	for position, value := range values {
		builder.arguments = append(builder.arguments, value)
		fragment = strings.ReplaceAll(fragment, "@"+strconv.Itoa(position+1), "$"+strconv.Itoa(len(builder.arguments)))
	}
	return strings.TrimSpace(fragment)
}

func (builder *searchQueryBuilder) addClause(fragment string, values ...any) {
	builder.clauses = append(builder.clauses, builder.bind(fragment, values...))
}

func (builder *searchQueryBuilder) addScope(query FileSearchQuery) {
	switch query.Scope {
	case SearchScopeDescendants:
		builder.addClause(queries.SearchScopeDescendantsQuery, query.ScopePath)
	case SearchScopeChildren:
		builder.addClause(queries.SearchScopeChildrenQuery, query.ScopePath)
	}
}

func (builder *searchQueryBuilder) addNameTerms(searchText string) bool {
	termPatterns, hasTerms := utils.BuildSearchTermPatterns(searchText)
	if !hasTerms {
		return false
	}
	builder.addClause(queries.SearchFilterNameTermsQuery, termPatterns.DrivingPattern, pq.Array(termPatterns.AllPatterns))
	return true
}

func (builder *searchQueryBuilder) addKinds(kinds []FileSearchKind) {
	if len(kinds) == 0 {
		return
	}
	alternatives := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		alternatives = append(alternatives, builder.kindAlternative(kind))
	}
	builder.clauses = append(builder.clauses, "("+strings.Join(alternatives, " OR ")+")")
}

func (builder *searchQueryBuilder) kindAlternative(kind FileSearchKind) string {
	if kind == SearchKindFolder {
		return builder.bind(queries.SearchFilterKindFolderQuery)
	}
	if kind == SearchKindOther {
		return builder.bind(queries.SearchFilterKindOtherQuery, pq.Array(knownFileFormats()))
	}
	return builder.bind(queries.SearchFilterKindFormatQuery, pq.Array(formatsOfKind(kind)))
}

func formatsOfKind(kind FileSearchKind) []string {
	switch kind {
	case SearchKindDocument:
		return utils.DocumentFormats
	case SearchKindImage:
		return utils.ImageFormats
	case SearchKindAudio:
		return utils.AudioFormats
	case SearchKindVideo:
		return utils.VideoFormats
	case SearchKindArchive:
		return utils.ArchiveFormats
	}
	return nil
}

func knownFileFormats() []string {
	knownFormats := []string{}
	for _, kind := range []FileSearchKind{SearchKindDocument, SearchKindImage, SearchKindAudio, SearchKindVideo, SearchKindArchive} {
		knownFormats = append(knownFormats, formatsOfKind(kind)...)
	}
	return knownFormats
}

func (builder *searchQueryBuilder) addFilter(filter FileSearchFilter) {
	builder.addKinds(filter.Kinds)
	if filter.ModifiedFrom != nil {
		builder.addClause(queries.SearchFilterModifiedFromQuery, *filter.ModifiedFrom)
	}
	if filter.ModifiedTo != nil {
		builder.addClause(queries.SearchFilterModifiedBeforeQuery, filter.ModifiedTo.AddDate(0, 0, 1))
	}
	if filter.MinSize != nil {
		builder.addClause(queries.SearchFilterMinSizeQuery, *filter.MinSize)
	}
	if filter.MaxSize != nil {
		builder.addClause(queries.SearchFilterMaxSizeQuery, *filter.MaxSize)
	}
	switch filter.Tier {
	case TierHot:
		builder.addClause(queries.SearchFilterTierHotQuery)
	case TierCold:
		builder.addClause(queries.SearchFilterTierColdQuery)
	}
	if filter.OnlyStarred {
		builder.addClause(queries.SearchFilterStarredQuery)
	}
}

func (builder *searchQueryBuilder) orderFragment(searchText string, filter FileSearchFilter) (string, error) {
	switch filter.Sort {
	case SearchSortRelevance, "":
		normalizedText := strings.Join(utils.SplitSearchTerms(searchText), " ")
		return builder.bind(queries.SearchOrderRelevanceQuery, normalizedText, utils.BuildPrefixLikePattern(normalizedText), utils.BuildContainsLikePattern(normalizedText)), nil
	case SearchSortName:
		return pickSearchOrder(filter.Order, SearchOrderAscending, queries.SearchOrderNameAscQuery, queries.SearchOrderNameDescQuery), nil
	case SearchSortSize:
		return pickSearchOrder(filter.Order, SearchOrderDescending, queries.SearchOrderSizeAscQuery, queries.SearchOrderSizeDescQuery), nil
	case SearchSortModified:
		return pickSearchOrder(filter.Order, SearchOrderDescending, queries.SearchOrderModifiedAscQuery, queries.SearchOrderModifiedDescQuery), nil
	}
	return "", fmt.Errorf("unsupported file search sort: %q", filter.Sort)
}

func pickSearchOrder(requestedOrder FileSearchOrder, defaultOrder FileSearchOrder, ascendingFragment string, descendingFragment string) string {
	if requestedOrder == "" {
		requestedOrder = defaultOrder
	}
	if requestedOrder == SearchOrderDescending {
		return descendingFragment
	}
	return ascendingFragment
}

func buildSearchFilesQuery(query FileSearchQuery) (string, []any, error) {
	builder := &searchQueryBuilder{}
	if !builder.addNameTerms(query.Query) {
		return "", nil, fmt.Errorf("file search needs at least one term")
	}
	builder.addScope(query)
	builder.addFilter(query.Filter)

	orderFragment, err := builder.orderFragment(query.Query, query.Filter)
	if err != nil {
		return "", nil, err
	}
	pageFragment := builder.bind(queries.SearchPageQuery, query.PageSize+1, utils.CalculateOffset(query.Page, query.PageSize))

	var assembled strings.Builder
	assembled.WriteString(queries.SearchSelectQuery)
	for _, clause := range builder.clauses {
		assembled.WriteString("\n    AND ")
		assembled.WriteString(clause)
	}
	assembled.WriteString("\n")
	assembled.WriteString(orderFragment)
	assembled.WriteString("\n")
	assembled.WriteString(pageFragment)
	return assembled.String(), builder.arguments, nil
}

func (r *Repository) SearchActiveFiles(query FileSearchQuery) (utils.PaginationResponse[FileModel], error) {
	statement, arguments, err := buildSearchFilesQuery(query)
	if err != nil {
		return utils.PaginationResponse[FileModel]{}, err
	}
	return r.runFilesPageQuery(statement, query.Page, query.PageSize, arguments)
}
