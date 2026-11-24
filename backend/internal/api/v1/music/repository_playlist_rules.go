package music

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type playlistRowScanner interface {
	Scan(dest ...any) error
}

func scanPlaylist(rowScanner playlistRowScanner) (PlaylistModel, error) {
	var playlist PlaylistModel
	var encodedRules []byte

	err := rowScanner.Scan(
		&playlist.ID, &playlist.Name, &playlist.Description,
		&playlist.IsSystem, &playlist.CreatedAt, &playlist.UpdatedAt,
		&playlist.TrackCount, &playlist.IsAIGenerated,
		&playlist.Kind, &encodedRules, &playlist.TotalLengthSeconds,
	)
	if err != nil {
		return playlist, err
	}

	if encodedRules != nil {
		var rules PlaylistRules
		if err := json.Unmarshal(encodedRules, &rules); err != nil {
			return playlist, fmt.Errorf("regras da playlist %d invalidas: %w", playlist.ID, err)
		}
		playlist.Rules = &rules
	}
	return playlist, nil
}

func (r *Repository) CreateRulePlaylist(tx *sql.Tx, name string, description string, rules PlaylistRules) (PlaylistModel, error) {
	playlist := PlaylistModel{
		Name:        name,
		Description: description,
		Kind:        PlaylistKindRule,
		Rules:       &rules,
	}

	encodedRules, err := json.Marshal(rules)
	if err != nil {
		return playlist, fmt.Errorf("falha ao codificar regras da playlist: %w", err)
	}

	err = tx.QueryRow(queries.CreateRulePlaylistQuery, name, description, string(encodedRules)).Scan(
		&playlist.ID, &playlist.CreatedAt, &playlist.UpdatedAt,
	)
	if err != nil {
		return playlist, fmt.Errorf("falha ao criar playlist por regra: %w", err)
	}

	return playlist, nil
}

func (r *Repository) UpdatePlaylistRules(tx *sql.Tx, id int, rules PlaylistRules) error {
	encodedRules, err := json.Marshal(rules)
	if err != nil {
		return fmt.Errorf("falha ao codificar regras da playlist: %w", err)
	}

	result, err := tx.Exec(queries.UpdatePlaylistRulesQuery, string(encodedRules), id)
	if err != nil {
		return fmt.Errorf("falha ao atualizar regras da playlist: %w", err)
	}

	updatedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updatedRows == 0 {
		return fmt.Errorf("playlist por regra nao encontrada: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Repository) CountPlaylistRuleTracks(rules PlaylistRules, playlistID *int) (PlaylistRuleTotals, error) {
	var totals PlaylistRuleTotals

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		totalsByPlaylistID, err := queryPlaylistRuleTotals(tx, []playlistRuleQuerySpec{buildPlaylistRuleQuerySpec(rules, playlistID)})
		if err != nil {
			return err
		}
		totals = totalsByPlaylistID[playlistIDKey(playlistID)]
		return nil
	})
	if err != nil {
		return totals, fmt.Errorf("falha ao contar faixas da regra: %w", err)
	}

	return totals, nil
}

func playlistIDKey(playlistID *int) int {
	if playlistID == nil {
		return 0
	}
	return *playlistID
}

func queryPlaylistRuleTotals(tx *sql.Tx, specs []playlistRuleQuerySpec) (map[int]PlaylistRuleTotals, error) {
	encodedSpecs, err := encodePlaylistRuleQuerySpecs(specs)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(queries.CountPlaylistRuleTracksQuery, pq.Array(utils.AudioFormats), encodedSpecs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalsByPlaylistID := make(map[int]PlaylistRuleTotals, len(specs))
	for rows.Next() {
		var playlistID sql.NullInt64
		var totals PlaylistRuleTotals
		if err := rows.Scan(&playlistID, &totals.TrackCount, &totals.TotalLengthSeconds); err != nil {
			return nil, err
		}
		totalsByPlaylistID[int(playlistID.Int64)] = totals
	}
	return totalsByPlaylistID, rows.Err()
}

func applyPlaylistRuleTotals(tx *sql.Tx, playlists []PlaylistModel) error {
	specs := make([]playlistRuleQuerySpec, 0)
	for _, playlist := range playlists {
		if playlist.Kind != PlaylistKindRule || playlist.Rules == nil {
			continue
		}
		playlistID := playlist.ID
		specs = append(specs, buildPlaylistRuleQuerySpec(*playlist.Rules, &playlistID))
	}
	if len(specs) == 0 {
		return nil
	}

	totalsByPlaylistID, err := queryPlaylistRuleTotals(tx, specs)
	if err != nil {
		return err
	}

	for index := range playlists {
		totals, isRulePlaylist := totalsByPlaylistID[playlists[index].ID]
		if playlists[index].Kind != PlaylistKindRule || !isRulePlaylist {
			continue
		}
		playlists[index].TrackCount = totals.TrackCount
		playlists[index].TotalLengthSeconds = totals.TotalLengthSeconds
	}
	return nil
}

func scanPlaylistTrack(rowScanner playlistRowScanner) (PlaylistTrackModel, error) {
	var track PlaylistTrackModel
	err := rowScanner.Scan(

		&track.ID, &track.PlaylistID, &track.FileID,
		&track.Position, &track.AddedAt,
		&track.FileID, &track.FileName, &track.FilePath,
		&track.FileParentPath, &track.FileFormat, &track.FileSize,
		&track.FileUpdatedAt, &track.FileCreatedAt,
		&track.LastInteraction, &track.LastBackup,
		&track.FileType, &track.FileCheckSum,
		&track.FileDeletedAt, &track.FileStarred,
		&track.MetadataID, &track.MetadataFileId, &track.MetadataPath,
		&track.MetadataMime, &track.MetadataLength, &track.MetadataBitrate,
		&track.MetadataSampleRate, &track.MetadataChannels,
		&track.MetadataBitrateMode, &track.MetadataEncoderInfo, &track.MetadataBitDepth,
		&track.MetadataTitle, &track.MetadataArtist, &track.MetadataAlbum,
		&track.MetadataAlbumArtist, &track.MetadataTrackNumber, &track.MetadataGenre,
		&track.MetadataComposer, &track.MetadataYear, &track.MetadataRecordingDate,
		&track.MetadataEncoder, &track.MetadataPublisher, &track.MetadataOriginalReleaseDate,
		&track.MetadataOriginalArtist, &track.MetadataLyricist, &track.MetadataLyrics,
		&track.MetadataCreatedAt,
	)
	return track, err
}

func findPlaylistRuleSpec(tx *sql.Tx, playlistID int) (*playlistRuleQuerySpec, error) {
	var encodedRules []byte
	err := tx.QueryRow(queries.GetPlaylistRulesByIDQuery, playlistID).Scan(&encodedRules)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rules PlaylistRules
	if err := json.Unmarshal(encodedRules, &rules); err != nil {
		return nil, fmt.Errorf("regras da playlist %d invalidas: %w", playlistID, err)
	}
	ruleSpec := buildPlaylistRuleQuerySpec(rules, &playlistID)
	return &ruleSpec, nil
}

func (r *Repository) findPlaylistRuleSpec(playlistID int) (*playlistRuleQuerySpec, error) {
	var ruleSpec *playlistRuleQuerySpec
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		foundSpec, err := findPlaylistRuleSpec(tx, playlistID)
		ruleSpec = foundSpec
		return err
	})
	return ruleSpec, err
}

func isRulePlaylist(tx *sql.Tx, playlistID int) (bool, error) {
	ruleSpec, err := findPlaylistRuleSpec(tx, playlistID)
	return ruleSpec != nil, err
}

func refuseWriteOnRulePlaylist(tx *sql.Tx, playlistID int) error {
	isRule, err := isRulePlaylist(tx, playlistID)
	if err != nil {
		return err
	}
	if isRule {
		return ErrPlaylistRuleReadOnly
	}
	return nil
}

func excludeTrackFromRulePlaylist(tx *sql.Tx, playlistID int, fileID int) error {
	if _, err := tx.Exec(queries.AddPlaylistRuleExclusionQuery, playlistID, fileID); err != nil {
		return fmt.Errorf("falha ao excluir faixa da playlist por regra: %w", err)
	}
	return nil
}

func queryPlaylistTrackRows(tx *sql.Tx, playlistID int, limit int, offset int) (*sql.Rows, error) {
	ruleSpec, err := findPlaylistRuleSpec(tx, playlistID)
	if err != nil {
		return nil, err
	}
	if ruleSpec == nil {
		return tx.Query(queries.GetPlaylistTracksQuery, playlistID, limit, offset)
	}

	encodedSpecs, err := encodePlaylistRuleQuerySpecs([]playlistRuleQuerySpec{*ruleSpec})
	if err != nil {
		return nil, err
	}
	return tx.Query(queries.GetPlaylistRuleTracksQuery, pq.Array(utils.AudioFormats), encodedSpecs, limit, offset)
}
