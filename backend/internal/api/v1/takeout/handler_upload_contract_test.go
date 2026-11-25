package takeout

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

func decodeResponseFields(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &fields); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return fields
}

func assertExactKeys(t *testing.T, fields map[string]any, expectedKeys []string) {
	t.Helper()
	actualKeys := make([]string, 0, len(fields))
	for key := range fields {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)
	sort.Strings(expectedKeys)
	if !reflect.DeepEqual(actualKeys, expectedKeys) {
		t.Fatalf("expected keys %v, got %v", expectedKeys, actualKeys)
	}
}

func TestInitUploadResponseFieldNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{"file_name":"takeout.zip","size":10}`)
	service := &handlerServiceMock{initFn: func(InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
		return InitTakeoutUploadResultDto{UploadID: "u1", ChunkSize: 2048, ReceivedSize: 4, Resumed: true}, nil
	}}

	NewHandler(service, nil).InitUploadHandler(ctx)

	fields := decodeResponseFields(t, recorder)
	assertExactKeys(t, fields, []string{"upload_id", "chunk_size", "received_size", "resumed"})
	if fields["upload_id"] != "u1" || fields["chunk_size"] != float64(2048) || fields["received_size"] != float64(4) || fields["resumed"] != true {
		t.Fatalf("unexpected values %v", fields)
	}
}

func TestUploadChunkResponseFieldNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildChunkRequest(t, "u1", "0", true)

	NewHandler(&handlerServiceMock{}, nil).UploadChunkHandler(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	fields := decodeResponseFields(t, recorder)
	assertExactKeys(t, fields, []string{"received"})
	if fields["received"] != true {
		t.Fatalf("unexpected values %v", fields)
	}
}

func TestCompleteUploadResponseFieldNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)
	service := &handlerServiceMock{completeFn: func(CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
		return TakeoutImportResultDto{JobID: 12, ImportID: 5, Message: "accepted"}, nil
	}}

	NewHandler(service, nil).CompleteUploadHandler(ctx)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", recorder.Code)
	}
	fields := decodeResponseFields(t, recorder)
	assertExactKeys(t, fields, []string{"job_id", "import_id", "message"})
	if fields["job_id"] != float64(12) || fields["import_id"] != float64(5) || fields["message"] != "accepted" {
		t.Fatalf("unexpected values %v", fields)
	}
}
