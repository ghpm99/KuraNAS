package files

import (
	queries "nas-go/api/pkg/database/queries/files"
)

type ChildrenSortKey string

type SortDirection string

const (
	SortByName      ChildrenSortKey = "name"
	SortBySize      ChildrenSortKey = "size"
	SortByUpdatedAt ChildrenSortKey = "updated_at"
	SortByCreatedAt ChildrenSortKey = "created_at"

	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type ChildrenSort struct {
	Key       ChildrenSortKey
	Direction SortDirection
}

var DefaultChildrenSort = ChildrenSort{}

var childrenQueriesBySort = map[ChildrenSort]string{
	{SortByName, SortAscending}:       queries.GetChildrenSortedByNameAscQuery,
	{SortByName, SortDescending}:      queries.GetChildrenSortedByNameDescQuery,
	{SortBySize, SortAscending}:       queries.GetChildrenSortedBySizeAscQuery,
	{SortBySize, SortDescending}:      queries.GetChildrenSortedBySizeDescQuery,
	{SortByUpdatedAt, SortAscending}:  queries.GetChildrenSortedByUpdatedAtAscQuery,
	{SortByUpdatedAt, SortDescending}: queries.GetChildrenSortedByUpdatedAtDescQuery,
	{SortByCreatedAt, SortAscending}:  queries.GetChildrenSortedByCreatedAtAscQuery,
	{SortByCreatedAt, SortDescending}: queries.GetChildrenSortedByCreatedAtDescQuery,
}

// ParseChildrenSort converts the raw "sort" and "order" query values into a
// ChildrenSort. An unknown or empty sort key yields DefaultChildrenSort; a
// known key with an unknown or empty order sorts ascending.
func ParseChildrenSort(rawKey string, rawDirection string) ChildrenSort {
	requestedSort := ChildrenSort{Key: ChildrenSortKey(rawKey), Direction: SortDirection(rawDirection)}
	if requestedSort.Direction != SortDescending {
		requestedSort.Direction = SortAscending
	}
	if _, isKnown := childrenQueriesBySort[requestedSort]; !isKnown {
		return DefaultChildrenSort
	}
	return requestedSort
}

func (childrenSort ChildrenSort) childrenQuery() string {
	if childrenQuery, isKnown := childrenQueriesBySort[childrenSort]; isKnown {
		return childrenQuery
	}
	return queries.GetChildrenByParentPathQuery
}
