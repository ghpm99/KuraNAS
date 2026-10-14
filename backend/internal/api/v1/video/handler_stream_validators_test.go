package video

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const streamFixtureSize = "36"

func (fixture *streamAccessFixture) streamWithHeaders(headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	request.RemoteAddr = "10.0.0.1:1234"
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	fixture.router.ServeHTTP(recorder, request)
	return recorder
}

func TestStreamVideoServesPartialContentForValidRange(t *testing.T) {
	recorder := newStreamAccessFixture(t).streamWithHeaders(map[string]string{"Range": "bytes=2-5"})
	if recorder.Code != http.StatusPartialContent || recorder.Body.String() != "2345" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Range") != "bytes 2-5/"+streamFixtureSize {
		t.Fatalf("unexpected Content-Range %q", recorder.Header().Get("Content-Range"))
	}
	if recorder.Header().Get("Accept-Ranges") != "bytes" || recorder.Header().Get("Cache-Control") != "public, max-age=3600" {
		t.Fatalf("missing caching headers: %v", recorder.Header())
	}
}

func TestStreamVideoRespondsRangeNotSatisfiableWithoutBody(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	for _, rangeHeader := range []string{"bytes=1000-", "bytes=36-40", "bytes=-0"} {
		recorder := fixture.streamWithHeaders(map[string]string{"Range": rangeHeader})
		if recorder.Code != http.StatusRequestedRangeNotSatisfiable {
			t.Fatalf("%s: unexpected status %d", rangeHeader, recorder.Code)
		}
		if recorder.Header().Get("Content-Range") != "bytes */"+streamFixtureSize || recorder.Body.Len() != 0 {
			t.Fatalf("%s: unexpected Content-Range %q body %q", rangeHeader, recorder.Header().Get("Content-Range"), recorder.Body.String())
		}
	}
	fixture.assertCounts(t, 0, 0, 0)
}

func TestStreamVideoIgnoresMalformedRange(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	for _, rangeHeader := range []string{"bytes=abc", "bytes=5-2", "items=0-1", "bytes=", "bytes=--3"} {
		recorder := fixture.streamWithHeaders(map[string]string{"Range": rangeHeader})
		if recorder.Code != http.StatusOK || recorder.Body.Len() != 36 {
			t.Fatalf("%s: unexpected response %d with %d bytes", rangeHeader, recorder.Code, recorder.Body.Len())
		}
	}
}

func TestStreamVideoServesOnlyFirstRangeOfMultiRange(t *testing.T) {
	recorder := newStreamAccessFixture(t).streamWithHeaders(map[string]string{"Range": "bytes=0-1,5-6"})
	if recorder.Code != http.StatusPartialContent || recorder.Body.String() != "01" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Range") != "bytes 0-1/"+streamFixtureSize {
		t.Fatalf("unexpected Content-Range %q", recorder.Header().Get("Content-Range"))
	}
}

func TestStreamVideoSendsValidatorsOnFullAndPartialResponses(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	for _, rangeHeader := range []string{"", "bytes=0-3"} {
		headers := map[string]string{}
		if rangeHeader != "" {
			headers["Range"] = rangeHeader
		}
		recorder := fixture.streamWithHeaders(headers)
		if recorder.Header().Get("ETag") == "" || recorder.Header().Get("Last-Modified") == "" {
			t.Fatalf("range %q: missing validators: %v", rangeHeader, recorder.Header())
		}
	}
}

func TestStreamVideoIfRangeDecidesBetweenPartialAndFullContent(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	validators := fixture.streamWithHeaders(nil)
	eTag := validators.Header().Get("ETag")
	lastModified := validators.Header().Get("Last-Modified")

	cases := []struct {
		name       string
		ifRange    string
		rangeValue string
		wantStatus int
	}{
		{"matching etag", eTag, "bytes=2-5", http.StatusPartialContent},
		{"different etag", `"other"`, "bytes=2-5", http.StatusOK},
		{"matching date", lastModified, "bytes=2-5", http.StatusPartialContent},
		{"different date", time.Unix(1000, 0).UTC().Format(http.TimeFormat), "bytes=2-5", http.StatusOK},
		{"different etag with unsatisfiable range", `"other"`, "bytes=1000-", http.StatusOK},
		{"matching etag with unsatisfiable range", eTag, "bytes=1000-", http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range cases {
		recorder := fixture.streamWithHeaders(map[string]string{"Range": testCase.rangeValue, "If-Range": testCase.ifRange})
		if recorder.Code != testCase.wantStatus {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.wantStatus, recorder.Code)
		}
		if testCase.wantStatus == http.StatusOK && recorder.Body.Len() != 36 {
			t.Fatalf("%s: expected full body, got %d bytes", testCase.name, recorder.Body.Len())
		}
	}
}

func TestStreamVideoRespondsNotModifiedWhenValidatorsMatch(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	validators := fixture.streamWithHeaders(nil)

	byETag := fixture.streamWithHeaders(map[string]string{"If-None-Match": validators.Header().Get("ETag")})
	if byETag.Code != http.StatusNotModified || byETag.Body.Len() != 0 {
		t.Fatalf("expected 304 by etag, got %d with %d bytes", byETag.Code, byETag.Body.Len())
	}
	byDate := fixture.streamWithHeaders(map[string]string{"If-Modified-Since": validators.Header().Get("Last-Modified")})
	if byDate.Code != http.StatusNotModified {
		t.Fatalf("expected 304 by date, got %d", byDate.Code)
	}
	stale := fixture.streamWithHeaders(map[string]string{"If-None-Match": `"other"`})
	if stale.Code != http.StatusOK {
		t.Fatalf("expected 200 for different etag, got %d", stale.Code)
	}
}
