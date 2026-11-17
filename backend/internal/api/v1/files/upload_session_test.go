package files

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestIsValidUploadIDAcceptsOnlyGeneratedFormat(t *testing.T) {
	generatedID, err := generateUploadID()
	if err != nil || !isValidUploadID(generatedID) {
		t.Fatalf("generated id %q must be valid (%v)", generatedID, err)
	}
	rejectedIDs := []string{"", "../x", "..", "a/b", strings.Repeat("A", 32), strings.Repeat("a", 31), strings.Repeat("a", 33), strings.Repeat("a", 31) + "/", strings.Repeat("g", 32)}
	for _, rejectedID := range rejectedIDs {
		if isValidUploadID(rejectedID) {
			t.Fatalf("id %q must be rejected", rejectedID)
		}
	}
}

func TestLoadUploadSessionRejectsMalformedID(t *testing.T) {
	newUploadTestEnvironment(t, nil)

	_, err := loadUploadSession("../x")

	requireOperationError(t, err, http.StatusBadRequest, "ERROR_UPLOAD_ID_INVALID")
}

func TestLoadUploadSessionUnknownIDIsNotFound(t *testing.T) {
	newUploadTestEnvironment(t, nil)

	_, err := loadUploadSession(strings.Repeat("a", 32))

	requireOperationError(t, err, http.StatusNotFound, "ERROR_UPLOAD_SESSION_NOT_FOUND")
}

func TestLoadUploadSessionCorruptedMetaIsInternalError(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("b", 32)
	if err := createUploadSession(UploadSession{UploadID: uploadID}); err != nil {
		t.Fatalf("createUploadSession failed: %v", err)
	}
	if err := os.WriteFile(uploadSessionMetaPath(uploadID), []byte("{"), 0644); err != nil {
		t.Fatalf("corrupt meta: %v", err)
	}

	_, err := loadUploadSession(uploadID)

	requireOperationError(t, err, http.StatusInternalServerError, "ERROR_UPLOAD_FAILED")
}

func TestCreateAndSaveUploadSessionRejectInvalidID(t *testing.T) {
	newUploadTestEnvironment(t, nil)

	if err := createUploadSession(UploadSession{UploadID: "../x"}); err == nil {
		t.Fatalf("createUploadSession must reject invalid id")
	}
	if err := saveUploadSession(UploadSession{UploadID: "../x"}); err == nil {
		t.Fatalf("saveUploadSession must reject invalid id")
	}
}

func TestDeleteUploadSessionRejectsMalformedID(t *testing.T) {
	newUploadTestEnvironment(t, nil)

	requireOperationError(t, deleteUploadSession("../x"), http.StatusBadRequest, "ERROR_UPLOAD_ID_INVALID")
}

func TestUploadSessionsRootFallsBackNextToEntryPoint(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	setCapturesPathForTest(t, "")

	if !strings.HasPrefix(uploadSessionsRootDir(), environment.entryPoint) {
		t.Fatalf("fallback root %q must live under the entry point", uploadSessionsRootDir())
	}
}

func TestFindOpenUploadSessionSkipsGarbageEntries(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	if err := os.MkdirAll(uploadSessionDir(strings.Repeat("c", 32)), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(uploadSessionsRootDir()+"/not-an-id", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if _, found := findOpenUploadSession(UploadSession{FileName: "a"}); found {
		t.Fatalf("garbage entries must never match")
	}
}
