package music

import (
	"context"
	"database/sql"
	"log"
	"sort"
	"strings"

	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
)

// RebuildAIClusters regenerates the AI-curated playlists from the music library.
// It clusters only artists not seen before (incremental, so the model is asked
// about each artist once), persists the artist -> category mapping, prunes
// artists that left the library, and materializes one real, flagged playlist per
// category. Safe to call repeatedly: a run with no new artists only refreshes
// track membership of the existing playlists.
func (s *Service) RebuildAIClusters(ctx context.Context) error {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return err
	}

	inputs, entriesByArtist := buildArtistClusterInputs(indexEntries)

	persisted, err := s.Repository.GetArtistClusters()
	if err != nil {
		return err
	}

	knownKeys := make(map[string]bool, len(inputs))
	for _, item := range inputs {
		knownKeys[item.Key] = true
	}

	// Seed the mapping with persisted assignments that still exist in the
	// library; the rest are stale and get pruned later.
	mapping := make(map[string]string, len(persisted))
	for _, cluster := range persisted {
		if knownKeys[cluster.ArtistKey] {
			mapping[cluster.ArtistKey] = cluster.ClusterName
		}
	}

	unclustered := make([]artistClusterInput, 0)
	for _, item := range inputs {
		if _, ok := mapping[item.Key]; !ok {
			unclustered = append(unclustered, item)
		}
	}

	if len(unclustered) > 0 && s.AIService != nil {
		assignments := s.clusterArtists(ctx, unclustered, distinctClusterNames(mapping))
		if err := s.persistAssignments(assignments); err != nil {
			return err
		}
		for _, assignment := range assignments {
			mapping[assignment.ArtistKey] = assignment.ClusterName
		}
	}

	if len(mapping) == 0 {
		return nil
	}

	if err := s.pruneArtistClusters(mapping); err != nil {
		return err
	}

	return s.materializeClusterPlaylists(mapping, entriesByArtist)
}

// clusterArtists runs the model over the unclustered artists in small batches,
// threading the running set of category names so later batches reuse them. A
// failed batch is logged and skipped; its artists simply retry on the next run.
func (s *Service) clusterArtists(ctx context.Context, unclustered []artistClusterInput, existingNames []string) []clusterAssignment {
	names := append([]string(nil), existingNames...)
	all := make([]clusterAssignment, 0, len(unclustered))

	for _, batch := range batchArtists(unclustered, defaultArtistClusterBatchSize) {
		assignments, err := s.clusterBatch(ctx, batch, names)
		if err != nil {
			log.Printf("AI artist clustering batch failed: %v\n", err)
			continue
		}
		all = append(all, assignments...)
		names = mergeClusterNames(names, assignments)
	}

	return all
}

func (s *Service) clusterBatch(ctx context.Context, batch []artistClusterInput, existingNames []string) ([]clusterAssignment, error) {
	// No per-request deadline here: the single source of truth for how long an
	// AI call may take is the provider's own HTTP timeout, configured at runtime
	// in the ai_providers table (Settings → AI Providers). Local/cloud models are
	// slow, so a hardcoded ceiling here only fought that config.
	prompt := prompts.MusicArtistClustersUserPrompt(
		defaultMaxNewClustersPerBatch,
		formatExistingClusters(existingNames),
		formatArtistsForPrompt(batch),
	)

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskClassification,
		SystemPrompt: prompts.MusicArtistClustersSystemPrompt(),
		Prompt:       prompt,
		MaxTokens:    800,
		Temperature:  0.2,
	})
	if err != nil {
		return nil, err
	}

	parsed, err := parseClusterResponse(resp.Content)
	if err != nil {
		return nil, err
	}

	return assignBatchClusters(batch, parsed), nil
}

func (s *Service) persistAssignments(assignments []clusterAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, assignment := range assignments {
			if err := s.Repository.UpsertArtistCluster(tx, ArtistClusterModel{
				ArtistKey:   assignment.ArtistKey,
				Artist:      assignment.Artist,
				ClusterName: assignment.ClusterName,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) pruneArtistClusters(mapping map[string]string) error {
	keys := make([]string, 0, len(mapping))
	for key := range mapping {
		keys = append(keys, key)
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeleteArtistClustersExcept(tx, keys)
	})
}

// materializeClusterPlaylists projects the artist -> category mapping onto real
// playlist rows: it creates a flagged playlist per category, refreshes its
// tracks, and deletes AI playlists whose category no longer exists.
func (s *Service) materializeClusterPlaylists(mapping map[string]string, entriesByArtist map[string][]MusicLibraryIndexEntryModel) error {
	clusterTracks := buildClusterTrackIDs(mapping, entriesByArtist)

	existingPlaylists, err := s.Repository.GetAIPlaylists()
	if err != nil {
		return err
	}

	byName := make(map[string]PlaylistModel, len(existingPlaylists))
	for _, playlist := range existingPlaylists {
		byName[strings.ToLower(playlist.Name)] = playlist
	}

	names := make([]string, 0, len(clusterTracks))
	for name := range clusterTracks {
		names = append(names, name)
	}
	sort.Strings(names)

	wanted := make(map[string]bool, len(names))

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, name := range names {
			wanted[strings.ToLower(name)] = true

			playlist, ok := byName[strings.ToLower(name)]
			if !ok {
				created, createErr := s.Repository.CreateAIPlaylist(tx, name, "")
				if createErr != nil {
					return createErr
				}
				playlist = created
			}

			if err := s.Repository.ReplacePlaylistTracks(tx, playlist.ID, clusterTracks[name]); err != nil {
				return err
			}
		}

		for _, playlist := range existingPlaylists {
			if wanted[strings.ToLower(playlist.Name)] {
				continue
			}
			if err := s.Repository.DeletePlaylist(tx, playlist.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

// buildArtistClusterInputs aggregates the library into one record per artist,
// carrying a track count and a representative genre hint (the artist's most
// common genre tag). It also returns the per-artist entries used to materialize
// each playlist's tracks.
func buildArtistClusterInputs(indexEntries []MusicLibraryIndexEntryModel) ([]artistClusterInput, map[string][]MusicLibraryIndexEntryModel) {
	type accumulator struct {
		artist     string
		trackCount int
		genres     map[string]int
	}

	accumulators := map[string]*accumulator{}
	entriesByArtist := map[string][]MusicLibraryIndexEntryModel{}

	for _, entry := range indexEntries {
		artist := preferredArtist(entry)
		if artist == "" {
			continue
		}

		key := normalizeLookupKey(artist)
		acc := accumulators[key]
		if acc == nil {
			acc = &accumulator{artist: artist, genres: map[string]int{}}
			accumulators[key] = acc
		}

		acc.trackCount++
		for _, genre := range normalizeGenreLabels(entry.Genre) {
			acc.genres[genre]++
		}
		entriesByArtist[key] = append(entriesByArtist[key], entry)
	}

	inputs := make([]artistClusterInput, 0, len(accumulators))
	for key, acc := range accumulators {
		inputs = append(inputs, artistClusterInput{
			Key:        key,
			Artist:     acc.artist,
			GenreHint:  topGenre(acc.genres),
			TrackCount: acc.trackCount,
		})
	}

	sort.Slice(inputs, func(left, right int) bool {
		if inputs[left].TrackCount != inputs[right].TrackCount {
			return inputs[left].TrackCount > inputs[right].TrackCount
		}
		return inputs[left].Artist < inputs[right].Artist
	})

	return inputs, entriesByArtist
}

// topGenre returns the most frequent genre, breaking ties lexicographically so
// the hint is deterministic.
func topGenre(genres map[string]int) string {
	best := ""
	bestCount := 0
	for genre, count := range genres {
		if count > bestCount || (count == bestCount && best != "" && genre < best) {
			best = genre
			bestCount = count
		}
	}
	return best
}

// distinctClusterNames lists the unique category names currently in the mapping,
// sorted for stable prompts.
func distinctClusterNames(mapping map[string]string) []string {
	seen := map[string]bool{}
	names := make([]string, 0, len(mapping))
	for _, name := range mapping {
		trimmed := normalizeText(name)
		if trimmed == "" || seen[strings.ToLower(trimmed)] {
			continue
		}
		seen[strings.ToLower(trimmed)] = true
		names = append(names, trimmed)
	}
	sort.Strings(names)
	return names
}

// buildClusterTrackIDs groups artists by their assigned category and produces
// the ordered, de-duplicated file IDs for each playlist. Tracks are ordered by
// artist, then album, then track number, so an artist plays as a contiguous run
// rather than being scattered across the playlist.
func buildClusterTrackIDs(mapping map[string]string, entriesByArtist map[string][]MusicLibraryIndexEntryModel) map[string][]int {
	clusterEntries := map[string][]MusicLibraryIndexEntryModel{}
	for artistKey, cluster := range mapping {
		name := normalizeText(cluster)
		if name == "" {
			continue
		}
		clusterEntries[name] = append(clusterEntries[name], entriesByArtist[artistKey]...)
	}

	result := make(map[string][]int, len(clusterEntries))
	for name, entries := range clusterEntries {
		sortGenreTracks(entries)
		result[name] = uniqueFileIDs(entries)
	}

	return result
}

func uniqueFileIDs(entries []MusicLibraryIndexEntryModel) []int {
	ids := make([]int, 0, len(entries))
	seen := map[int]bool{}
	for _, entry := range entries {
		if seen[entry.FileID] {
			continue
		}
		seen[entry.FileID] = true
		ids = append(ids, entry.FileID)
	}
	return ids
}
