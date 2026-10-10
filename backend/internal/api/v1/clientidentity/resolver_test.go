package clientidentity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func resolveFor(t *testing.T, target string, headers map[string]string) (string, bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.RemoteAddr = "10.1.2.3:5555"
	for name, headerValue := range headers {
		request.Header.Set(name, headerValue)
	}
	context.Request = request
	return Resolve(context)
}

func TestResolvePrefersHeaderOverQuery(t *testing.T) {
	clientID, isValid := resolveFor(t, "/x?client_id=query-client-1", map[string]string{Header: "header-client-1"})
	if !isValid || clientID != "header-client-1" {
		t.Fatalf("got %q valid=%v", clientID, isValid)
	}
}

func TestResolveFallsBackToQuery(t *testing.T) {
	clientID, isValid := resolveFor(t, "/x?client_id=query-client-1", nil)
	if !isValid || clientID != "query-client-1" {
		t.Fatalf("got %q valid=%v", clientID, isValid)
	}
}

func TestResolveFallsBackToClientIP(t *testing.T) {
	clientID, isValid := resolveFor(t, "/x", nil)
	if !isValid || clientID != "10.1.2.3" {
		t.Fatalf("got %q valid=%v", clientID, isValid)
	}
}

func TestResolveRejectsMalformedIDs(t *testing.T) {
	for _, malformedID := range []string{"short", "has space in it 123", "bad_chars_underscore", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		if _, isValid := resolveFor(t, "/x", map[string]string{Header: malformedID}); isValid {
			t.Fatalf("expected %q to be invalid", malformedID)
		}
	}
}
