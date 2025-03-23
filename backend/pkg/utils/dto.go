package utils

type Pagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	HasNext  bool `json:"has_next"`
	HasPrev  bool `json:"has_prev"`
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) GetHasPrev() bool {
	return p.Page > 1
}

type Optional[T any] struct {
	Value    T
	HasValue bool
}

type PaginationResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}
