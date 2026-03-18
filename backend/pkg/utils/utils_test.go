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

type simpleFilterFixture struct {
	ID      int       `filter:"id"`
	Name    string    `filter:"name"`
	Enabled bool      `filter:"enabled"`
	Date    time.Time `filter:"date"`
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
	got = replacePlaceholder("$4", "SELECT $4", int64(9))
	if !strings.Contains(got, "9") {
		t.Fatalf("expected int64 replacement, got: %s", got)
	}
	got = replacePlaceholder("$5", "SELECT $5", struct{}{})
	if !strings.Contains(got, "NULL") {
		t.Fatalf("expected default NULL replacement, got: %s", got)
	}

}

func TestGenerateFilterFromContextAndParseContextQuery(t *testing.T) {
	filter := filterFixture{}
	ctx := newGinContext("id=7&name=alice&enabled=true&date=2025-02-03")
	GenerateFilterFromContext(ctx, &filter)

	if !filter.ID.HasValue || filter.ID.Value != 7 {
		t.Fatalf("expected optional id=7, got %+v", filter.ID)
	}
	if !filter.Name.HasValue || filter.Name.Value != "alice" {
		t.Fatalf("expected optional name=alice, got %+v", filter.Name)
	}
	if !filter.Enabled.HasValue || !filter.Enabled.Value {
		t.Fatalf("expected optional enabled=true, got %+v", filter.Enabled)
	}
	if !filter.Date.HasValue || filter.Date.Value.Format("2006-01-02") != "2025-02-03" {
		t.Fatalf("expected optional date parsed, got %+v", filter.Date)
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
	var tm time.Time
	parseContextQuery(reflect.TypeOf(tm), reflect.ValueOf(&tm).Elem(), "2025-02-03")
	if tm.Format("2006-01-02") != "2025-02-03" {
		t.Fatalf("expected parsed time, got %v", tm)
	}

	optInt := Optional[int]{}
	parseContextQuery(reflect.TypeOf(optInt), reflect.ValueOf(&optInt).Elem(), "11")
	if !optInt.HasValue || optInt.Value != 11 {
		t.Fatalf("expected optional int parsed, got %+v", optInt)
	}

	optBool := Optional[bool]{}
	parseContextQuery(reflect.TypeOf(optBool), reflect.ValueOf(&optBool).Elem(), "true")
	if !optBool.HasValue || !optBool.Value {
		t.Fatalf("expected optional bool parsed, got %+v", optBool)
	}
}

func TestGenerateFilterFromContextOptionalMissing(t *testing.T) {
	filter := filterFixture{
		ID:      NewOptional(10),
		Name:    NewOptional("x"),
		Enabled: NewOptional(true),
		Date:    NewOptional(time.Now()),
	}

	ctx := newGinContext("")
	GenerateFilterFromContext(ctx, &filter)

	if filter.ID.HasValue || filter.Name.HasValue || filter.Enabled.HasValue || filter.Date.HasValue {
		t.Fatalf("expected optional fields to be reset when query params are missing")
	}
}

func TestGenerateFilterFromContextWithPrimitiveFields(t *testing.T) {
	filter := simpleFilterFixture{}
	ctx := newGinContext("id=8&name=bob&enabled=true&date=2026-01-03")
	GenerateFilterFromContext(ctx, &filter)

	if filter.ID != 8 || filter.Name != "bob" || !filter.Enabled {
		t.Fatalf("unexpected primitive filter result: %+v", filter)
	}
	if filter.Date.Format("2006-01-02") != "2026-01-03" {
		t.Fatalf("unexpected parsed date in filter: %v", filter.Date)
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

	p2 := PaginationResponse[int]{
		Items: []int{1},
		Pagination: Pagination{
			Page:     1,
			PageSize: 5,
		},
	}
	p2.SetHasNext()
	if p2.Pagination.HasNext {
		t.Fatalf("expected has next false")
	}

	if (&Pagination{Page: 1}).GetHasPrev() {
		t.Fatalf("expected false for first page")
	}
	if !(&Pagination{Page: 2}).GetHasPrev() {
		t.Fatalf("expected true for page > 1")
	}
}

func TestCalculateOffset(t *testing.T) {
	if got := CalculateOffset(0, 0); got != 0 {
		t.Fatalf("expected offset 0, got %d", got)
	}
	if got := CalculateOffset(3, 10); got != 20 {
		t.Fatalf("expected offset 20, got %d", got)
	}
}

func TestGetFormatTypeByExtension(t *testing.T) {
	cases := []struct {
		ext      string
		expected string
	}{
		{".MP3", FormatTypeAudio},
		{".wav", FormatTypeAudio},
		{".aac", FormatTypeAudio},
		{".flac", FormatTypeAudio},
		{".png", FormatTypeImage},
		{".gif", FormatTypeImage},
		{".svg", FormatTypeImage},
		{".webp", FormatTypeImage},
		{".webm", FormatTypeVideo},
		{".ogg", FormatTypeVideo},
		{".mov", FormatTypeVideo},
		{".pdf", FormatTypeDocument},
		{".txt", FormatTypeDocument},
		{".html", FormatTypeDocument},
		{".xml", FormatTypeDocument},
		{".json", FormatTypeDocument},
		{".csv", FormatTypeDocument},
		{".zip", FormatTypeArchive},
		{".rar", FormatTypeArchive},
		{".7z", FormatTypeArchive},
		{".tar", FormatTypeArchive},
		{".gz", FormatTypeArchive},
		{".unknown", FormatTypeUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.ext, func(t *testing.T) {
			if got := GetFormatTypeByExtension(tc.ext); got.Type != tc.expected {
				t.Fatalf("expected %v for %s, got %v", tc.expected, tc.ext, got.Type)
			}
		})
	}
}

func TestRunPythonScriptAndStructHelpers(t *testing.T) {
	if _, err := RunPythonScript(""); err == nil {
		t.Fatalf("expected empty script name error")
	}
	if _, err := RunPythonScript(ScriptType("missing_script.py"), "/tmp/file"); err == nil {
		t.Fatalf("expected execution error for missing script")
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

	args = StructToArgs(&obj)
	if len(args) != 2 || args[1].(string) != "x" {
		t.Fatalf("unexpected StructToArgs pointer output: %#v", args)
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
	if _, err := GetFileChecksum(filepath.Join(tmpDir, "missing")); err == nil {
		t.Fatalf("expected GetFileChecksum error for missing file")
	}
}
