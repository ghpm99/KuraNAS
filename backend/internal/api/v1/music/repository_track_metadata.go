package music

import (
	"database/sql"
	"fmt"
	"strconv"

	queries "nas-go/api/pkg/database/queries/music"

	"github.com/lib/pq"
)

// TrackMetadataPatchGroup is one patch applied to every track in FileIDs.
type TrackMetadataPatchGroup struct {
	FileIDs []int
	Patch   TrackMetadataPatch
}

// ApplyTrackMetadataPatch writes the patch as the user correction of every
// active indexed track in fileIDs, inside tx. Fields absent from the patch keep
// their current correction; fields present but empty are cleared back to the
// tag. It then drops corrections left with no field, recomputes the catalog
// grouping keys and reconciles album groupings with compilationArtistLabel.
func (r *AudioMetadataRepository) ApplyTrackMetadataPatch(tx *sql.Tx, fileIDs []int, patch TrackMetadataPatch, compilationArtistLabel string) error {
	return r.ApplyTrackMetadataPatchGroups(tx, []TrackMetadataPatchGroup{{FileIDs: fileIDs, Patch: patch}}, compilationArtistLabel)
}

// ApplyTrackMetadataPatchGroups writes each group's patch as the user
// correction of its tracks, then drops empty corrections, recomputes the
// catalog keys of every touched track and reconciles album groupings once,
// all inside tx.
func (r *AudioMetadataRepository) ApplyTrackMetadataPatchGroups(tx *sql.Tx, groups []TrackMetadataPatchGroup, compilationArtistLabel string) error {
	touchedFileIDs := []int{}
	for _, group := range groups {
		if _, err := tx.Exec(queries.UpsertAudioMetadataOverrideQuery, buildOverrideUpsertArgs(group.FileIDs, group.Patch)...); err != nil {
			return fmt.Errorf("falha ao gravar correcao de metadados: %w", err)
		}
		touchedFileIDs = append(touchedFileIDs, group.FileIDs...)
	}
	if _, err := tx.Exec(queries.DeleteEmptyAudioMetadataOverridesQuery, pq.Array(touchedFileIDs)); err != nil {
		return fmt.Errorf("falha ao remover correcoes vazias: %w", err)
	}
	if err := r.RecomputeCatalogKeys(tx, touchedFileIDs); err != nil {
		return err
	}
	_, err := r.reconcileAlbumGroupingsInTx(tx, compilationArtistLabel)
	return err
}

func buildOverrideUpsertArgs(fileIDs []int, patch TrackMetadataPatch) []any {
	return []any{
		pq.Array(fileIDs),
		patch.Title.IsSet, patch.Title.Value,
		patch.Artist.IsSet, patch.Artist.Value,
		patch.Album.IsSet, patch.Album.Value,
		patch.AlbumArtist.IsSet, patch.AlbumArtist.Value,
		patch.Genres.IsSet, patch.Genres.joinedGenres(),
		patch.Year.IsSet, yearAsText(patch.Year),
		patch.DiscNumber.IsSet, patch.DiscNumber.Value,
		patch.TrackNumber.IsSet, patch.TrackNumber.Value,
	}
}

func yearAsText(year optionalInteger) *string {
	if year.Value == nil {
		return nil
	}
	yearText := strconv.Itoa(*year.Value)
	return &yearText
}
