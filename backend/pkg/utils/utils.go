package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ParseInt(value string, context *gin.Context) int {
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
	}

	return valueInt
}

func ParseDate(value string, context *gin.Context) time.Time {
	valueTime, err := time.Parse("2006-01-02", value)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
	}
	return valueTime
}

func PrintQuery(query string, args []interface{}) {

	finalQuery := query

	for i := len(args) - 1; i >= 0; i-- {
		placeHolder := fmt.Sprintf("$%s", strconv.Itoa(i+1))
		finalQuery = replacePlaceholder(placeHolder, finalQuery, args[i])
	}

	fmt.Println("Query gerada:", finalQuery)
}

func replacePlaceholder(placeHolder string, query string, param interface{}) string {
	var value string

	switch v := param.(type) {
	case string:
		value = fmt.Sprintf("'%s'", v)
	case int, int64, float64:
		value = fmt.Sprintf("%v", v)
	case bool:
		value = "TRUE"
	case time.Time:
		value = v.Format("2006-01-02")
	default:
		fmt.Println("value", v)
		value = "NULL"
	}
	return strings.ReplaceAll(query, placeHolder, value)
}

func GenerateFilterFromContext[T any](context *gin.Context, filter *T) {
	t := reflect.TypeOf(filter).Elem()
	v := reflect.ValueOf(filter).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("filter")

		if tag == "" {
			continue
		}

		paramValue := context.Query(tag)
		fieldValue := v.Field(i)
		fieldType := field.Type

		if paramValue == "" {
			if fieldType.Kind() == reflect.Struct && fieldType.Name() == "Optional" {
				fieldValue.Set(reflect.ValueOf(Optional[any]{
					HasValue: false,
					Value:    nil,
				}))
			}
			continue
		}

		switch fieldType.Kind() {
		case reflect.Int:
			if intValue, err := strconv.Atoi(paramValue); err == nil {

				fieldValue.SetInt(int64(intValue))
			}
		case reflect.String:

			fieldValue.SetString(paramValue)
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(paramValue); err == nil {

				fieldValue.SetBool(boolValue)
			}
		case reflect.Struct:
			// Verifica se é um `time.Time`
			fmt.Println(fieldType, paramValue)
			if fieldType == reflect.TypeOf(time.Time{}) {
				if parsedTime, err := time.Parse("2006-01-02", paramValue); err == nil {
					fmt.Println(parsedTime)
					fieldValue.Set(reflect.ValueOf(parsedTime))
				}
			}
		default:
			// Verifica se é um `Optional`
			if fieldType.Kind() == reflect.Struct && fieldType.Name() == "Optional" {
				elemType := fieldType.Field(0).Type // Tipo genérico do `Optional`

				switch elemType.Kind() {
				case reflect.Int:
					if intValue, err := strconv.Atoi(paramValue); err == nil {

						fieldValue.Set(reflect.ValueOf(NewOptional(intValue)))
					}
				case reflect.String:

					fieldValue.Set(reflect.ValueOf(NewOptional(paramValue)))
				case reflect.Bool:
					if boolValue, err := strconv.ParseBool(paramValue); err == nil {

						fieldValue.Set(reflect.ValueOf(NewOptional(boolValue)))
					}
				case reflect.Struct:
					// Para `Optional[time.Time]`
					if elemType == reflect.TypeOf(time.Time{}) {
						if parsedTime, err := time.Parse("2006-01-02", paramValue); err == nil {

							fieldValue.Set(reflect.ValueOf(NewOptional(parsedTime)))
						}
					}
				}
			}

		}

	}

}

func parseContextQuery(fieldType reflect.Type, fieldValue reflect.Value, paramValue string) {
	switch fieldType.Kind() {
	case reflect.Int:
		if intValue, err := strconv.Atoi(paramValue); err == nil {

			fieldValue.SetInt(int64(intValue))
		}
	case reflect.String:

		fieldValue.SetString(paramValue)
	case reflect.Bool:
		if boolValue, err := strconv.ParseBool(paramValue); err == nil {

			fieldValue.SetBool(boolValue)
		}
	case reflect.Struct:
		// Verifica se é um `time.Time`
		if fieldType == reflect.TypeOf(time.Time{}) {
			if parsedTime, err := time.Parse("2006-01-02", paramValue); err == nil {

				fieldValue.Set(reflect.ValueOf(parsedTime))
			}
		}
	default:
		// Verifica se é um `Optional`
		if fieldType.Kind() == reflect.Struct && fieldType.Name() == "Optional" {
			elemType := fieldType.Field(0).Type // Tipo genérico do `Optional`

			switch elemType.Kind() {
			case reflect.Int:
				if intValue, err := strconv.Atoi(paramValue); err == nil {

					fieldValue.Set(reflect.ValueOf(NewOptional(intValue)))
				}
			case reflect.String:

				fieldValue.Set(reflect.ValueOf(NewOptional(paramValue)))
			case reflect.Bool:
				if boolValue, err := strconv.ParseBool(paramValue); err == nil {

					fieldValue.Set(reflect.ValueOf(NewOptional(boolValue)))
				}
			case reflect.Struct:
				// Para `Optional[time.Time]`
				if elemType == reflect.TypeOf(time.Time{}) {
					if parsedTime, err := time.Parse("2006-01-02", paramValue); err == nil {

						fieldValue.Set(reflect.ValueOf(NewOptional(parsedTime)))
					}
				}
			}
		}

	}
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{Value: value, HasValue: true}
}

func (p *PaginationResponse[T]) SetHasNext() {
	if len(p.Items) <= p.Pagination.PageSize {
		p.Pagination.HasNext = false
		return
	}

	p.Items = p.Items[:len(p.Items)-1]
	p.Pagination.HasNext = true
}

func (p *PaginationResponse[T]) SetHasPrev() {
	p.Pagination.HasPrev = p.Pagination.Page > 1
}

func (p *PaginationResponse[T]) UpdatePagination() {
	p.SetHasNext()
	p.SetHasPrev()
}
