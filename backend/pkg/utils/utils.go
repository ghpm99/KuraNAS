package utils

import (
	"fmt"
	"net/http"
	"os/exec"
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

func CalculateOffset(page int, pageSize int) int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	return (page - 1) * pageSize
}

const (
	FormatTypeImage    = "image"
	FormatTypeAudio    = "audio"
	FormatTypeVideo    = "video"
	FormatTypeDocument = "document"
	FormatTypeArchive  = "archive"
	FormatTypeUnknown  = "unknown"
)

type FormatType struct {
	Type        string
	Mime        string
	Description string
}

func GetFormatTypeByExtension(ext string) FormatType {
	ext = strings.ToLower(ext)
	switch ext {
	// Imagens
	case ".jpg", ".jpeg":
		return FormatType{Type: FormatTypeImage, Mime: "image/jpeg", Description: "IMAGE_JPEG"}
	case ".png":
		return FormatType{Type: FormatTypeImage, Mime: "image/png", Description: "IMAGE_PNG"}
	case ".gif":
		return FormatType{Type: FormatTypeImage, Mime: "image/gif", Description: "IMAGE_GIF"}
	case ".bmp":
		return FormatType{Type: FormatTypeImage, Mime: "image/bmp", Description: "IMAGE_BMP"}
	case ".svg":
		return FormatType{Type: FormatTypeImage, Mime: "image/svg+xml", Description: "IMAGE_SVG"}
	case ".webp":
		return FormatType{Type: FormatTypeImage, Mime: "image/webp", Description: "IMAGE_WEBP"}

	// Áudios
	case ".mp3":
		return FormatType{Type: FormatTypeAudio, Mime: "audio/mpeg", Description: "AUDIO_MP3"}
	case ".wav":
		return FormatType{Type: FormatTypeAudio, Mime: "audio/wav", Description: "AUDIO_WAV"}
	case ".aac":
		return FormatType{Type: FormatTypeAudio, Mime: "audio/aac", Description: "AUDIO_AAC"}
	case ".flac":
		return FormatType{Type: FormatTypeAudio, Mime: "audio/flac", Description: "AUDIO_FLAC"}

	// Vídeos
	case ".mp4":
		return FormatType{Type: FormatTypeVideo, Mime: "video/mp4", Description: "VIDEO_MP4"}
	case ".webm":
		return FormatType{Type: FormatTypeVideo, Mime: "video/webm", Description: "VIDEO_WEBM"}
	case ".ogg":
		return FormatType{Type: FormatTypeVideo, Mime: "video/ogg", Description: "VIDEO_OGG"}
	case ".mov":
		return FormatType{Type: FormatTypeVideo, Mime: "video/quicktime", Description: "VIDEO_MOV"}

	// Documentos
	case ".pdf":
		return FormatType{Type: FormatTypeDocument, Mime: "application/pdf", Description: "DOCUMENT_PDF"}
	case ".txt":
		return FormatType{Type: FormatTypeDocument, Mime: "text/plain", Description: "DOCUMENT_TXT"}
	case ".html", ".htm":
		return FormatType{Type: FormatTypeDocument, Mime: "text/html", Description: "DOCUMENT_HTML"}
	case ".xml":
		return FormatType{Type: FormatTypeDocument, Mime: "application/xml", Description: "DOCUMENT_XML"}
	case ".json":
		return FormatType{Type: FormatTypeDocument, Mime: "application/json", Description: "DOCUMENT_JSON"}
	case ".csv":
		return FormatType{Type: FormatTypeDocument, Mime: "text/csv", Description: "DOCUMENT_CSV"}

	// Outros
	case ".zip":
		return FormatType{Type: FormatTypeArchive, Mime: "application/zip", Description: "ARCHIVE_ZIP"}
	case ".rar":
		return FormatType{Type: FormatTypeArchive, Mime: "application/vnd.rar", Description: "ARCHIVE_RAR"}
	case ".7z":
		return FormatType{Type: FormatTypeArchive, Mime: "application/x-7z-compressed", Description: "ARCHIVE_7Z"}
	case ".tar":
		return FormatType{Type: FormatTypeArchive, Mime: "application/x-tar", Description: "ARCHIVE_TAR"}
	case ".gz":
		return FormatType{Type: FormatTypeArchive, Mime: "application/gzip", Description: "ARCHIVE_GZIP"}

	default:
		return FormatType{Type: FormatTypeUnknown, Mime: "", Description: "UNKNOWN_FORMAT"}
	}
}

const (
	ImageMetadata = "image_metadata.py"
	AudioMetadata = "audio_metadata.py"
	VideoMetadata = "video_metadata.py"
)

func RunPythonScript(scriptName string, arg ...string) (string, error) {
	args := append([]string{"scripts/" + scriptName}, arg...)
	cmd := exec.Command("scripts/.venv/bin/python", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao executar script python: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func StructToArgs(v interface{}) []interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	args := make([]interface{}, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		args[i] = val.Field(i).Interface()
	}
	return args
}

func StructToScanPtrs(v interface{}) []interface{} {
	val := reflect.ValueOf(v).Elem()
	ptrs := make([]interface{}, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		ptrs[i] = val.Field(i).Addr().Interface()
	}
	return ptrs
}
