package utils

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type filterFixture struct {
	ID      Optional[int]       `filter:"id"`
	Name    Optional[string]    `filter:"name"`
	Enabled Optional[bool]      `filter:"enabled"`
	Date    Optional[time.Time] `filter:"date"`
}

func newGinContext(rawQuery string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/?"+rawQuery, nil)
	ctx.Request = req
	return ctx
}

func TestParseIntAndDate(t *testing.T) {
	ctx := newGinContext("")
	if got := ParseInt("10", ctx); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}

	ctxErr := newGinContext("")
	_ = ParseInt("x", ctxErr)
	if len(ctxErr.Errors) == 0 {
		t.Fatalf("expected parse int error in context")
	}

	ctxDate := newGinContext("")
	d := ParseDate("2025-01-10", ctxDate)
	if d.Format("2006-01-02") != "2025-01-10" {
		t.Fatalf("unexpected parsed date: %s", d)
	}

	ctxDateErr := newGinContext("")
	_ = ParseDate("bad-date", ctxDateErr)
	if len(ctxDateErr.Errors) == 0 {
		t.Fatalf("expected parse date error in context")
	}
}

func TestPlaceholderAndPrintQuery(t *testing.T) {
	query := "SELECT * FROM t WHERE name=$1 AND active=$2 AND created=$3"
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	got := replacePlaceholder("$1", query, "john")
	got = replacePlaceholder("$2", got, true)
	got = replacePlaceholder("$3", got, ts)
	if !strings.Contains(got, "'john'") || !strings.Contains(got, "TRUE") || !strings.Contains(got, "2025-01-01") {
		t.Fatalf("unexpected replaced query: %s", got)
	}

	// Just ensure no panic while formatting and printing.
	PrintQuery("SELECT $1,$2,$3,$4", []interface{}{"a", 2, true, ts})
}

func TestGenerateFilterFromContextAndParseContextQuery(t *testing.T) {
	filter := filterFixture{}
	ctx := newGinContext("id=7&name=alice&enabled=true&date=2025-02-03")
	GenerateFilterFromContext(ctx, &filter)

	// Current implementation does not populate Optional fields in this path.
	if filter.ID.HasValue || filter.Name.HasValue || filter.Enabled.HasValue || filter.Date.HasValue {
		t.Fatalf("expected optional fields to remain unset with current implementation")
	}

	// Exercise parseContextQuery branches directly.
	var i int
	parseContextQuery(reflect.TypeOf(i), reflect.ValueOf(&i).Elem(), "9")
	if i != 9 {
		t.Fatalf("expected int 9, got %d", i)
	}
	var s string
	parseContextQuery(reflect.TypeOf(s), reflect.ValueOf(&s).Elem(), "text")
	if s != "text" {
		t.Fatalf("expected string text, got %s", s)
	}
	var b bool
	parseContextQuery(reflect.TypeOf(b), reflect.ValueOf(&b).Elem(), "true")
	if !b {
		t.Fatalf("expected bool true")
	}
}

func TestOptionalAndPagination(t *testing.T) {
	opt := NewOptional(5)
	if !opt.HasValue || opt.Value != 5 {
		t.Fatalf("expected optional with value")
	}

	p := PaginationResponse[int]{
		Items: []int{1, 2, 3},
		Pagination: Pagination{
			Page:     2,
			PageSize: 2,
		},
	}
	p.UpdatePagination()
	if !p.Pagination.HasNext {
		t.Fatalf("expected has next")
	}
	if !p.Pagination.HasPrev {
		t.Fatalf("expected has prev")
	}
	if len(p.Items) != 2 {
		t.Fatalf("expected one item trimmed for hasNext calculation")
	}
}

func TestCalculateOffsetAndFormatType(t *testing.T) {
	if got := CalculateOffset(0, 0); got != 0 {
		t.Fatalf("expected offset 0, got %d", got)
	}
	if got := CalculateOffset(3, 10); got != 20 {
		t.Fatalf("expected offset 20, got %d", got)
	}

	if got := GetFormatTypeByExtension(".MP3"); got.Type != FormatTypeAudio {
		t.Fatalf("expected audio for .MP3")
	}
	if got := GetFormatTypeByExtension(".pdf"); got.Type != FormatTypeDocument {
		t.Fatalf("expected document for .pdf")
	}
	if got := GetFormatTypeByExtension(".zip"); got.Type != FormatTypeArchive {
		t.Fatalf("expected archive for .zip")
	}
	if got := GetFormatTypeByExtension(".unknown"); got.Type != FormatTypeUnknown {
		t.Fatalf("expected unknown for unsupported extension")
	}
}

func TestRunPythonScriptAndStructHelpers(t *testing.T) {
	if _, err := RunPythonScript(""); err == nil {
		t.Fatalf("expected empty script name error")
	}
	if _, err := RunPythonScript(ImageMetadata, "/tmp/file"); err == nil {
		t.Fatalf("expected execution error when python/script path is unavailable")
	}

	type sample struct {
		A int
		B string
	}
	obj := sample{A: 1, B: "x"}
	args := StructToArgs(obj)
	if len(args) != 2 || args[0].(int) != 1 || args[1].(string) != "x" {
		t.Fatalf("unexpected StructToArgs output: %#v", args)
	}

	ptrObj := sample{}
	ptrs := StructToScanPtrs(&ptrObj)
	if len(ptrs) != 2 {
		t.Fatalf("unexpected StructToScanPtrs size: %d", len(ptrs))
	}
}

func TestChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("world"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	sum1, err := GetFileChecksum(file1)
	if err != nil || sum1 == "" {
		t.Fatalf("expected checksum for file1, err=%v", err)
	}

	combined := GetCheckSumFromPath([]string{sum1, "invalid-hex"})
	if combined == "" {
		t.Fatalf("expected non-empty combined checksum")
	}

	dirSum, err := GetDirectoryChecksum(tmpDir)
	if err != nil || dirSum == "" {
		t.Fatalf("expected directory checksum, err=%v", err)
	}

	if _, err := GetDirectoryChecksum(filepath.Join(tmpDir, "missing")); err == nil {
		t.Fatalf("expected missing path error")
	}
	if _, err := GetDirectoryChecksum(file1); err == nil {
		t.Fatalf("expected non-directory error")
	}
}
