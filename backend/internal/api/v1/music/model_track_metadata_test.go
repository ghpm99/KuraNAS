package music

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func decodePatch(t *testing.T, body string) TrackMetadataPatch {
	t.Helper()
	var patch TrackMetadataPatch
	if err := json.Unmarshal([]byte(body), &patch); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return patch
}

func TestOptionalTextDistinguishesAbsentNullEmptyAndValue(t *testing.T) {
	absent := decodePatch(t, `{}`).Title
	if absent.IsSet || absent.Value != nil {
		t.Fatalf("absent = %+v", absent)
	}

	for _, clearingBody := range []string{`{"title": null}`, `{"title": ""}`, `{"title": "   "}`} {
		cleared := decodePatch(t, clearingBody).Title
		if !cleared.IsSet || cleared.Value != nil {
			t.Fatalf("%s decoded to %+v, want set with nil value", clearingBody, cleared)
		}
	}

	valued := decodePatch(t, `{"title": "  Song  "}`).Title
	if !valued.IsSet || valued.Value == nil || *valued.Value != "Song" {
		t.Fatalf("valued = %+v", valued)
	}
}

func TestOptionalIntegerDistinguishesAbsentNullAndValue(t *testing.T) {
	absent := decodePatch(t, `{}`).Year
	if absent.IsSet || absent.Value != nil {
		t.Fatalf("absent = %+v", absent)
	}

	cleared := decodePatch(t, `{"year": null}`).Year
	if !cleared.IsSet || cleared.Value != nil {
		t.Fatalf("null = %+v", cleared)
	}

	valued := decodePatch(t, `{"year": 1999}`).Year
	if !valued.IsSet || valued.Value == nil || *valued.Value != 1999 {
		t.Fatalf("valued = %+v", valued)
	}
}

func TestOptionalIntegerRejectsNonIntegerValues(t *testing.T) {
	for _, body := range []string{`{"year": 1999.5}`, `{"year": "1999"}`, `{"disc_number": true}`} {
		var patch TrackMetadataPatch
		if err := json.Unmarshal([]byte(body), &patch); err == nil {
			t.Fatalf("%s must be refused", body)
		}
	}
}

func TestOptionalTextListDistinguishesAbsentNullEmptyAndValues(t *testing.T) {
	absent := decodePatch(t, `{}`).Genres
	if absent.IsSet || absent.Values != nil {
		t.Fatalf("absent = %+v", absent)
	}

	for _, clearingBody := range []string{`{"genres": null}`, `{"genres": []}`, `{"genres": ["", "  "]}`} {
		cleared := decodePatch(t, clearingBody).Genres
		if !cleared.IsSet || len(cleared.Values) != 0 || cleared.joinedGenres() != nil {
			t.Fatalf("%s decoded to %+v, want set and empty", clearingBody, cleared)
		}
	}

	valued := decodePatch(t, `{"genres": [" Rock ", "", "Pop"]}`).Genres
	if !valued.IsSet || !slices.Equal(valued.Values, []string{"Rock", "Pop"}) {
		t.Fatalf("valued = %+v", valued)
	}
	if joinedGenres := valued.joinedGenres(); joinedGenres == nil || *joinedGenres != "Rock;Pop" {
		t.Fatalf("joined genres = %v", joinedGenres)
	}
}

func TestOptionalTextListRejectsNonTextElements(t *testing.T) {
	var patch TrackMetadataPatch
	if err := json.Unmarshal([]byte(`{"genres": [1]}`), &patch); err == nil {
		t.Fatal("non text genre must be refused")
	}
}

func TestPatchTypeErrorNamesTheRefusedField(t *testing.T) {
	var patch TrackMetadataPatch
	err := json.Unmarshal([]byte(`{"title": "ok", "disc_number": "two"}`), &patch)

	var typeError *json.UnmarshalTypeError
	if !errors.As(err, &typeError) || typeError.Field != "disc_number" {
		t.Fatalf("expected type error on disc_number, got %v", err)
	}
}
