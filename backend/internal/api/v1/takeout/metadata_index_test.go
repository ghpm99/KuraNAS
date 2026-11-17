package takeout

import (
	"strings"
	"testing"
)

func metadataWithTimestamp(title string, timestamp string) TakeoutMetadata {
	return TakeoutMetadata{Title: title, PhotoTakenTime: takeoutTimestamp{Timestamp: timestamp}}
}

func assertFoundTimestamp(t *testing.T, found *TakeoutMetadata, expectedTimestamp string) {
	t.Helper()
	if found == nil {
		t.Fatalf("expected metadata with timestamp %q, found none", expectedTimestamp)
	}
	if found.PhotoTakenTime.Timestamp != expectedTimestamp {
		t.Fatalf("expected timestamp %q, got %q", expectedTimestamp, found.PhotoTakenTime.Timestamp)
	}
}

func TestMetadataIndexMatchesTheThreeJSONNameFormats(t *testing.T) {
	jsonNamesByTimestamp := map[string]string{
		"1": "A/foto.jpg.json",
		"2": "B/foto.jpg.supplemental-metadata.json",
		"3": "C/foto.jpg.supplemental-metad.json",
		"4": "D/foto.jpg.suppl.json",
		"5": "E/foto.jpg.s.json",
	}
	index := newMetadataIndex()
	for timestamp, jsonName := range jsonNamesByTimestamp {
		index.add(jsonName, metadataWithTimestamp("", timestamp))
	}

	for timestamp, jsonName := range jsonNamesByTimestamp {
		folder := jsonName[:1]
		assertFoundTimestamp(t, index.find(folder+"/foto.jpg"), timestamp)
	}
}

func TestMetadataIndexMatchesNameTruncatedTo46Characters(t *testing.T) {
	mediaName := strings.Repeat("a", 50) + ".jpg"
	truncatedStem := mediaName[:46]
	index := newMetadataIndex()
	index.add("A/"+truncatedStem+".json", metadataWithTimestamp("", "7"))

	assertFoundTimestamp(t, index.find("A/"+mediaName), "7")
}

func TestMetadataIndexTruncatedNameTiesAreBrokenByTitle(t *testing.T) {
	longPrefix := strings.Repeat("b", 46)
	firstMedia := longPrefix + "_first.jpg"
	secondMedia := longPrefix + "_second.jpg"
	index := newMetadataIndex()
	index.add("A/"+longPrefix+".json", metadataWithTimestamp(firstMedia, "10"))
	index.add("A/"+longPrefix+".supplemental-metadata.json", metadataWithTimestamp(secondMedia, "20"))

	assertFoundTimestamp(t, index.find("A/"+firstMedia), "10")
	assertFoundTimestamp(t, index.find("A/"+secondMedia), "20")
}

func TestMetadataIndexSeparatesSameMediaNameAcrossFolders(t *testing.T) {
	index := newMetadataIndex()
	index.add("A/IMG_0001.JPG.json", metadataWithTimestamp("IMG_0001.JPG", "111"))
	index.add("B/IMG_0001.JPG.supplemental-metadata.json", metadataWithTimestamp("IMG_0001.JPG", "222"))

	assertFoundTimestamp(t, index.find("A/IMG_0001.JPG"), "111")
	assertFoundTimestamp(t, index.find("B/IMG_0001.JPG"), "222")
	if unmatched := index.find("C/IMG_0001.JPG"); unmatched != nil {
		t.Fatalf("a folder without JSON must not borrow another folder's metadata, got %+v", unmatched)
	}
}

func TestMetadataIndexDoesNotMatchTruncatedNameOfAnotherTitle(t *testing.T) {
	index := newMetadataIndex()
	index.add("A/IMG.json", metadataWithTimestamp("IMG.png", "9"))

	if found := index.find("A/IMG_0001.jpg"); found != nil {
		t.Fatalf("a JSON titled for another file must not match by prefix, got %+v", found)
	}
}

func TestMetadataIndexPrefersExactNameOverTruncatedPrefix(t *testing.T) {
	index := newMetadataIndex()
	index.add("A/IMG_1.json", metadataWithTimestamp("", "1"))
	index.add("A/IMG_10.jpg.json", metadataWithTimestamp("IMG_10.jpg", "10"))

	assertFoundTimestamp(t, index.find("A/IMG_10.jpg"), "10")
}
