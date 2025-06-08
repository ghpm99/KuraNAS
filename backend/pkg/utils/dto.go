package utils

import (
	"database/sql"
	"fmt"
	"time"
)

type Pagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	HasNext  bool `json:"has_next"`
	HasPrev  bool `json:"has_prev"`
}

func (p *Pagination) GetHasPrev() bool {
	return p.Page > 1
}

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func (o *Optional[T]) IsEmpty() bool {
	return !o.HasValue
}
func (o *Optional[T]) Get() (T, error) {
	if !o.HasValue {
		return o.Value, fmt.Errorf("optional value is empty")
	}
	return o.Value, nil
}
func (o *Optional[T]) Set(value T) {
	o.Value = value
	o.HasValue = true
}

func (o *Optional[T]) ParseFromNullTime(value sql.NullTime) error {
	if !value.Valid {
		o.HasValue = false
		return nil
	}
	if _, ok := any(o.Value).(time.Time); !ok {
		return fmt.Errorf("value is not of type time.Time")
	}
	if value.Time.IsZero() {
		o.HasValue = false
		return nil
	}
	o.HasValue = true
	o.Value = any(value.Time).(T)
	return nil
}

func (o *Optional[T]) ParseToNullTime() (sql.NullTime, error) {
	if !o.HasValue {
		return sql.NullTime{Valid: false}, nil
	}
	value, ok := any(o.Value).(time.Time)
	if !ok {
		return sql.NullTime{}, fmt.Errorf("value is not of type time.Time")
	}
	return sql.NullTime{Time: value, Valid: true}, nil
}

type PaginationResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type TaskType int

const (
	ScanFiles      TaskType = 1
	ScanDir        TaskType = 2
	UpdateCheckSum TaskType = 3
)

type Task struct {
	Type TaskType
	Data string
}
