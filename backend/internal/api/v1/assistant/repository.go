package assistant

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/assistant"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(dbContext *database.DbContext) *Repository {
	return &Repository{DbContext: dbContext}
}

func (r *Repository) CreateConversation(title string) (ConversationModel, error) {
	var model ConversationModel
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.InsertConversationQuery, title).Scan(
			&model.ID, &model.Title, &model.CreatedAt, &model.UpdatedAt,
		)
	})
	if err != nil {
		return model, fmt.Errorf("CreateConversation: %w", err)
	}
	return model, nil
}

func (r *Repository) ConversationExists(id int) (bool, error) {
	var exists bool
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.ConversationExistsQuery, id).Scan(&exists)
	})
	if err != nil {
		return false, fmt.Errorf("ConversationExists: %w", err)
	}
	return exists, nil
}

func (r *Repository) ListConversations() ([]ConversationModel, error) {
	conversations := []ConversationModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListConversationsQuery)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var model ConversationModel
			if err := rows.Scan(&model.ID, &model.Title, &model.CreatedAt, &model.UpdatedAt); err != nil {
				return err
			}
			conversations = append(conversations, model)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListConversations: %w", err)
	}
	return conversations, nil
}

func (r *Repository) TouchConversation(id int) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.TouchConversationQuery, id)
		return err
	})
	if err != nil {
		return fmt.Errorf("TouchConversation: %w", err)
	}
	return nil
}

func (r *Repository) DeleteConversation(id int) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteConversationQuery, id)
		return err
	})
	if err != nil {
		return fmt.Errorf("DeleteConversation: %w", err)
	}
	return nil
}

func (r *Repository) AddMessage(conversationID int, role, content string) (MessageModel, error) {
	var model MessageModel
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.InsertMessageQuery, conversationID, role, content).Scan(
			&model.ID, &model.ConversationID, &model.Role, &model.Content, &model.CreatedAt,
		)
	})
	if err != nil {
		return model, fmt.Errorf("AddMessage: %w", err)
	}
	return model, nil
}

func (r *Repository) ListMessages(conversationID int) ([]MessageModel, error) {
	messages := []MessageModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListMessagesQuery, conversationID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var model MessageModel
			if err := rows.Scan(&model.ID, &model.ConversationID, &model.Role, &model.Content, &model.CreatedAt); err != nil {
				return err
			}
			messages = append(messages, model)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListMessages: %w", err)
	}
	return messages, nil
}
