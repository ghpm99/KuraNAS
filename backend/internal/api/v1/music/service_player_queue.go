package music

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrInvalidPlayerQueue = errors.New("invalid player queue")

func (s *Service) ReplacePlayerQueue(clientID string, request ReplacePlayerQueueRequest) error {
	if !request.isValid() {
		return ErrInvalidPlayerQueue
	}

	err := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.ReplacePlayerQueue(tx, clientID, request.FileIDs, request.CurrentIndex)
	})
	if err != nil {
		return fmt.Errorf("erro ao salvar a fila do player: %w", err)
	}
	return nil
}

func (s *Service) GetPlayerQueue(clientID string) (PlayerQueueDto, error) {
	entries, err := s.Repository.GetPlayerQueue(clientID)
	if err != nil {
		return PlayerQueueDto{}, err
	}
	currentIndex, err := s.Repository.GetPlayerQueueCurrentIndex(clientID)
	if err != nil {
		return PlayerQueueDto{}, err
	}

	items := make([]MusicQueueEntryDto, 0, len(entries))
	for _, entry := range entries {
		items = append(items, entry.ToDto())
	}
	return PlayerQueueDto{Items: items, CurrentIndex: currentIndex}, nil
}
