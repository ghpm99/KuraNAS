package music

import (
	"encoding/json"
	"sort"
	"strings"
)

const (
	// defaultArtistClusterBatchSize bounds how many artists a (possibly small,
	// local) model has to reason over in a single request. Keeping batches small
	// avoids long-list degradation and truncated/hallucinated output.
	defaultArtistClusterBatchSize = 40
	// defaultMaxNewClustersPerBatch caps how many brand-new categories the model
	// may invent per batch, nudging it to reuse the running cluster set.
	defaultMaxNewClustersPerBatch = 6
	// fallbackClusterName is the generic bucket for an artist the model never
	// classified and that carries no usable genre hint.
	fallbackClusterName = "Outros"
)

// artistClusterInput is the per-artist data fed to the clustering model. The
// genre hint is a weak fallback signal: the model is told to trust its own
// knowledge of the artist over it.
type artistClusterInput struct {
	Key        string
	Artist     string
	GenreHint  string
	TrackCount int
}

// aiClusterResponse mirrors the JSON contract the model must return.
type aiClusterResponse struct {
	Clusters []aiCluster `json:"clusters"`
}

type aiCluster struct {
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
}

// clusterAssignment is the validated mapping of one artist to one cluster.
type clusterAssignment struct {
	ArtistKey   string
	Artist      string
	ClusterName string
}

// batchArtists splits artists into fixed-size batches so the model never has to
// reason over a long list at once.
func batchArtists(inputs []artistClusterInput, batchSize int) [][]artistClusterInput {
	if batchSize < 1 {
		batchSize = defaultArtistClusterBatchSize
	}

	batches := make([][]artistClusterInput, 0, (len(inputs)+batchSize-1)/batchSize)
	for start := 0; start < len(inputs); start += batchSize {
		end := start + batchSize
		if end > len(inputs) {
			end = len(inputs)
		}
		batch := make([]artistClusterInput, end-start)
		copy(batch, inputs[start:end])
		batches = append(batches, batch)
	}

	return batches
}

// formatArtistsForPrompt renders one artist per line as "Artist [genre hint]",
// dropping the bracket when no hint is available.
func formatArtistsForPrompt(batch []artistClusterInput) string {
	lines := make([]string, 0, len(batch))
	for _, item := range batch {
		hint := normalizeText(item.GenreHint)
		if hint == "" {
			lines = append(lines, item.Artist)
			continue
		}
		lines = append(lines, item.Artist+" ["+hint+"]")
	}

	return strings.Join(lines, "\n")
}

// formatExistingClusters renders the running cluster names so the model can
// reuse them instead of inventing near-duplicates across batches.
func formatExistingClusters(names []string) string {
	cleaned := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		trimmed := normalizeText(name)
		if trimmed == "" || seen[strings.ToLower(trimmed)] {
			continue
		}
		seen[strings.ToLower(trimmed)] = true
		cleaned = append(cleaned, trimmed)
	}

	if len(cleaned) == 0 {
		return "(none yet)"
	}

	return strings.Join(cleaned, ", ")
}

// stripCodeFences removes Markdown ``` fences a model may wrap its JSON in.
func stripCodeFences(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	lines := strings.Split(trimmed, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		kept = append(kept, line)
	}

	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// parseClusterResponse decodes the model output into the cluster contract.
func parseClusterResponse(content string) (aiClusterResponse, error) {
	var parsed aiClusterResponse
	if err := json.Unmarshal([]byte(stripCodeFences(content)), &parsed); err != nil {
		return aiClusterResponse{}, err
	}

	return parsed, nil
}

// assignBatchClusters turns a raw model response into validated assignments for
// exactly the artists in the batch. It is the safety net around a small model:
//   - artists the model invented or echoed from outside the batch are dropped;
//   - an artist named in two clusters keeps the first;
//   - an artist the model forgot falls back to its genre hint, then to a generic
//     bucket, so every batch artist always lands somewhere meaningful.
func assignBatchClusters(batch []artistClusterInput, response aiClusterResponse) []clusterAssignment {
	byKey := make(map[string]artistClusterInput, len(batch))
	for _, item := range batch {
		byKey[item.Key] = item
	}

	assigned := make(map[string]string, len(batch))
	for _, cluster := range response.Clusters {
		name := normalizeText(cluster.Name)
		if name == "" {
			continue
		}
		for _, artist := range cluster.Artists {
			key := normalizeLookupKey(artist)
			if _, known := byKey[key]; !known {
				continue
			}
			if _, done := assigned[key]; done {
				continue
			}
			assigned[key] = name
		}
	}

	results := make([]clusterAssignment, 0, len(batch))
	for _, item := range batch {
		name, ok := assigned[item.Key]
		if !ok || name == "" {
			name = fallbackClusterFor(item)
		}
		results = append(results, clusterAssignment{
			ArtistKey:   item.Key,
			Artist:      item.Artist,
			ClusterName: name,
		})
	}

	sort.Slice(results, func(left, right int) bool {
		return results[left].Artist < results[right].Artist
	})

	return results
}

// fallbackClusterFor keeps an unclassified artist meaningful by using its genre
// hint, or a generic bucket when even that is missing.
func fallbackClusterFor(item artistClusterInput) string {
	if hint := normalizeGenreLabel(item.GenreHint); hint != "" {
		return hint
	}

	return fallbackClusterName
}

// mergeClusterNames appends newly seen cluster names to the running set,
// preserving order and ignoring case-insensitive duplicates.
func mergeClusterNames(existing []string, assignments []clusterAssignment) []string {
	seen := map[string]bool{}
	merged := make([]string, 0, len(existing)+len(assignments))
	for _, name := range existing {
		trimmed := normalizeText(name)
		if trimmed == "" || seen[strings.ToLower(trimmed)] {
			continue
		}
		seen[strings.ToLower(trimmed)] = true
		merged = append(merged, trimmed)
	}

	for _, assignment := range assignments {
		trimmed := normalizeText(assignment.ClusterName)
		if trimmed == "" || seen[strings.ToLower(trimmed)] {
			continue
		}
		seen[strings.ToLower(trimmed)] = true
		merged = append(merged, trimmed)
	}

	return merged
}
