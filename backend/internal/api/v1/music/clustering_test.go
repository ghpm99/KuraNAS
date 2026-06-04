package music

import (
	"reflect"
	"testing"
)

func clusterInput(artist string, genre string, trackCount int) artistClusterInput {
	return artistClusterInput{
		Key:        normalizeLookupKey(artist),
		Artist:     artist,
		GenreHint:  genre,
		TrackCount: trackCount,
	}
}

func TestBatchArtists(t *testing.T) {
	inputs := []artistClusterInput{
		clusterInput("A", "rock", 1),
		clusterInput("B", "rock", 1),
		clusterInput("C", "pop", 1),
		clusterInput("D", "pop", 1),
		clusterInput("E", "jazz", 1),
	}

	batches := batchArtists(inputs, 2)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if len(batches[0]) != 2 || len(batches[2]) != 1 {
		t.Fatalf("unexpected batch sizes: %d, %d", len(batches[0]), len(batches[2]))
	}

	// Mutating a batch must not affect the source slice (defensive copy).
	batches[0][0].Artist = "mutated"
	if inputs[0].Artist != "A" {
		t.Fatalf("batchArtists leaked the source slice")
	}
}

func TestBatchArtistsDefaultsAndEmpty(t *testing.T) {
	if batches := batchArtists(nil, 0); len(batches) != 0 {
		t.Fatalf("expected no batches for empty input, got %d", len(batches))
	}

	inputs := make([]artistClusterInput, defaultArtistClusterBatchSize+1)
	batches := batchArtists(inputs, 0)
	if len(batches) != 2 {
		t.Fatalf("expected fallback batch size to split into 2, got %d", len(batches))
	}
}

func TestFormatArtistsForPrompt(t *testing.T) {
	out := formatArtistsForPrompt([]artistClusterInput{
		clusterInput("The Beatles", "rock", 10),
		clusterInput("Unknown Local Band", "", 3),
	})

	expected := "The Beatles [rock]\nUnknown Local Band"
	if out != expected {
		t.Fatalf("unexpected prompt body:\n%q", out)
	}
}

func TestFormatExistingClusters(t *testing.T) {
	if out := formatExistingClusters(nil); out != "(none yet)" {
		t.Fatalf("expected placeholder for empty clusters, got %q", out)
	}

	out := formatExistingClusters([]string{"Classic Rock", "classic rock", " Metal ", ""})
	if out != "Classic Rock, Metal" {
		t.Fatalf("expected case-insensitive de-dup, got %q", out)
	}
}

func TestStripCodeFences(t *testing.T) {
	fenced := "```json\n{\"clusters\":[]}\n```"
	if got := stripCodeFences(fenced); got != "{\"clusters\":[]}" {
		t.Fatalf("fence not stripped: %q", got)
	}

	plain := "{\"clusters\":[]}"
	if got := stripCodeFences("  " + plain + "  "); got != plain {
		t.Fatalf("plain content altered: %q", got)
	}
}

func TestParseClusterResponse(t *testing.T) {
	parsed, err := parseClusterResponse("```\n{\"clusters\":[{\"name\":\"Rock\",\"artists\":[\"A\"]}]}\n```")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(parsed.Clusters) != 1 || parsed.Clusters[0].Name != "Rock" {
		t.Fatalf("unexpected parse result: %+v", parsed)
	}

	if _, err := parseClusterResponse("not json"); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestAssignBatchClustersHappyPath(t *testing.T) {
	batch := []artistClusterInput{
		clusterInput("The Beatles", "rock", 10),
		clusterInput("Nightwish", "rock", 8),
		clusterInput("Johnny Cash", "country", 5),
	}

	response := aiClusterResponse{Clusters: []aiCluster{
		{Name: "Classic Rock", Artists: []string{"The Beatles"}},
		{Name: "Symphonic Metal", Artists: []string{"Nightwish"}},
		{Name: "Country", Artists: []string{"Johnny Cash"}},
	}}

	got := assignBatchClusters(batch, response)
	want := map[string]string{
		normalizeLookupKey("The Beatles"): "Classic Rock",
		normalizeLookupKey("Nightwish"):   "Symphonic Metal",
		normalizeLookupKey("Johnny Cash"): "Country",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d assignments, got %d", len(want), len(got))
	}
	for _, assignment := range got {
		if want[assignment.ArtistKey] != assignment.ClusterName {
			t.Fatalf("artist %q assigned to %q, want %q", assignment.Artist, assignment.ClusterName, want[assignment.ArtistKey])
		}
	}
}

func TestAssignBatchClustersGuardrails(t *testing.T) {
	batch := []artistClusterInput{
		clusterInput("The Beatles", "rock", 10),
		clusterInput("Forgotten Artist", "samba", 4),
		clusterInput("No Hint Artist", "", 2),
	}

	response := aiClusterResponse{Clusters: []aiCluster{
		// Empty name is skipped entirely.
		{Name: "  ", Artists: []string{"The Beatles"}},
		// Hallucinated artist (not in batch) is dropped.
		{Name: "Pop", Artists: []string{"Taylor Swift"}},
		// Valid assignment, plus a duplicate of The Beatles in a second cluster.
		{Name: "Classic Rock", Artists: []string{"The Beatles"}},
		{Name: "Other Rock", Artists: []string{"The Beatles"}},
	}}

	got := assignBatchClusters(batch, response)
	byKey := map[string]string{}
	for _, assignment := range got {
		byKey[assignment.ArtistKey] = assignment.ClusterName
	}

	if byKey[normalizeLookupKey("The Beatles")] != "Classic Rock" {
		t.Fatalf("first cluster should win, got %q", byKey[normalizeLookupKey("The Beatles")])
	}
	if byKey[normalizeLookupKey("Forgotten Artist")] != "Samba" {
		t.Fatalf("forgotten artist should fall back to genre hint, got %q", byKey[normalizeLookupKey("Forgotten Artist")])
	}
	if byKey[normalizeLookupKey("No Hint Artist")] != fallbackClusterName {
		t.Fatalf("hintless artist should fall back to %q, got %q", fallbackClusterName, byKey[normalizeLookupKey("No Hint Artist")])
	}
}

func TestMergeClusterNames(t *testing.T) {
	existing := []string{"Classic Rock", "Metal"}
	assignments := []clusterAssignment{
		{ClusterName: "metal"},          // case-insensitive duplicate, ignored
		{ClusterName: "Country"},        // new
		{ClusterName: " Classic Rock "}, // trimmed duplicate, ignored
		{ClusterName: ""},               // empty, ignored
	}

	got := mergeClusterNames(existing, assignments)
	want := []string{"Classic Rock", "Metal", "Country"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected merge result: %#v", got)
	}
}
