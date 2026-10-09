package engine

import (
	"database/sql"
	"fmt"

	musicdom "nas-go/api/internal/api/v1/music"
)

const audioCatalogKeysBackfillPageSize = 500

func backfillAudioCatalogKeys(workerContext *WorkerContext) (int, error) {
	repository := workerContext.AudioMetadataRepository
	backfilledCount := 0
	afterAudioMetadataID := 0

	for {
		pendingRows, err := repository.ListAudioWithoutCatalogKeys(afterAudioMetadataID, audioCatalogKeysBackfillPageSize)
		if err != nil {
			return backfilledCount, fmt.Errorf("audio catalog keys backfill: list pending rows: %w", err)
		}
		if len(pendingRows) == 0 {
			return backfilledCount, nil
		}

		if err := persistAudioCatalogKeys(workerContext, pendingRows); err != nil {
			return backfilledCount, fmt.Errorf("audio catalog keys backfill: persist keys: %w", err)
		}

		backfilledCount += len(pendingRows)
		afterAudioMetadataID = pendingRows[len(pendingRows)-1].AudioMetadataID

		if len(pendingRows) < audioCatalogKeysBackfillPageSize {
			return backfilledCount, nil
		}
	}
}

func persistAudioCatalogKeys(workerContext *WorkerContext, pendingRows []musicdom.AudioCatalogKeySource) error {
	repository := workerContext.AudioMetadataRepository

	return repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		for _, pendingRow := range pendingRows {
			groupingKeys := musicdom.BuildCatalogGroupingKeys(pendingRow.Artist, pendingRow.AlbumArtist, pendingRow.Album, pendingRow.Genre)
			if err := repository.UpdateAudioCatalogKeys(tx, pendingRow.AudioMetadataID, groupingKeys); err != nil {
				return err
			}
		}
		return nil
	})
}
