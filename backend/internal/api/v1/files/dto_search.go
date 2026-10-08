package files

const MinSearchQueryLength = 2

type FileSearchParams struct {
	Query       string
	ParentID    int
	IsRecursive bool
	Page        int
	PageSize    int
}
