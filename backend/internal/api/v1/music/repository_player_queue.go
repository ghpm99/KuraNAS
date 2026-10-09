package music

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/music"

	"github.com/lib/pq"
)

func (r *Repository) ReplacePlayerQueue(tx *sql.Tx, clientID string, fileIDs []int, currentIndex int) error {
	if _, err := tx.Exec(queries.DeletePlayerQueueQuery, clientID); err != nil {
		return fmt.Errorf("falha ao limpar a fila do player: %w", err)
	}
	if _, err := tx.Exec(queries.InsertPlayerQueueQuery, clientID, pq.Array(fileIDs)); err != nil {
		return fmt.Errorf("falha ao gravar a fila do player: %w", err)
	}
	if _, err := tx.Exec(queries.UpsertPlayerQueueIndexQuery, clientID, currentIndex); err != nil {
		return fmt.Errorf("falha ao gravar a posicao da fila do player: %w", err)
	}
	return nil
}

func (r *Repository) GetPlayerQueue(clientID string) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetPlayerQueueQuery, clientID)
}

func (r *Repository) GetPlayerQueueCurrentIndex(clientID string) (int, error) {
	var currentIndex int
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetPlayerQueueCurrentIndexQuery, clientID).Scan(&currentIndex)
	})
	if err != nil {
		return 0, fmt.Errorf("falha ao obter a posicao da fila do player: %w", err)
	}
	return currentIndex, nil
}
