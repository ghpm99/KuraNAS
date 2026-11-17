package music

import (
	"fmt"

	queries "nas-go/api/pkg/database/queries/music"

	"github.com/lib/pq"
)

// FilterIndexedTrackFileIDs returns, ordered, the ids in fileIDs that are an
// active file with an indexed audio_metadata row.
func (r *AudioMetadataRepository) FilterIndexedTrackFileIDs(fileIDs []int) ([]int, error) {
	indexedFileIDs, err := queryFileIDs(r.Db, queries.FilterIndexedTrackFileIDsQuery, pq.Array(fileIDs))
	if err != nil {
		return nil, fmt.Errorf("falha ao filtrar faixas indexadas: %w", err)
	}
	return indexedFileIDs, nil
}

// ListTrackFileIDsByAlbumKey returns, ordered, every active indexed track whose
// effective catalog album key is albumKey.
func (r *AudioMetadataRepository) ListTrackFileIDsByAlbumKey(albumKey string) ([]int, error) {
	fileIDs, err := queryFileIDs(r.Db, queries.GetTrackFileIDsByAlbumKeyQuery, albumKey)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas do album: %w", err)
	}
	return fileIDs, nil
}

// ListTrackFileIDsByFolder returns, ordered, every active indexed track in
// folderPath or any of its subfolders.
func (r *AudioMetadataRepository) ListTrackFileIDsByFolder(folderPath string) ([]int, error) {
	fileIDs, err := queryFileIDs(r.Db, queries.GetTrackFileIDsByFolderQuery, folderPath, subfolderLikePattern(folderPath))
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas da pasta: %w", err)
	}
	return fileIDs, nil
}
