package watchfolders

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

func (s *Service) GetWatchFolders() ([]WatchFolderDto, error) {
	models, err := s.Repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("GetWatchFolders: %w", err)
	}

	dtos := make([]WatchFolderDto, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, model.ToDto())
	}
	return dtos, nil
}

func (s *Service) CreateWatchFolder(dto CreateWatchFolderDto) (WatchFolderDto, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(dto.Path))
	if err := s.validateWatchFolderPath(cleanPath, 0); err != nil {
		return WatchFolderDto{}, err
	}

	model := WatchFolderModel{
		Path:    cleanPath,
		Label:   strings.TrimSpace(dto.Label),
		Enabled: true,
	}

	var created WatchFolderModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		created, createErr = s.Repository.Create(tx, model)
		return createErr
	})
	if err != nil {
		return WatchFolderDto{}, fmt.Errorf("CreateWatchFolder: %w", err)
	}

	log.Println(i18n.Translate("WATCH_FOLDER_CREATED", created.Path))
	return created.ToDto(), nil
}

func (s *Service) UpdateWatchFolder(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error) {
	if id <= 0 {
		return WatchFolderDto{}, ErrInvalidWatchFolderID
	}

	existing, err := s.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WatchFolderDto{}, ErrWatchFolderNotFound
		}
		return WatchFolderDto{}, fmt.Errorf("UpdateWatchFolder get by id: %w", err)
	}

	updated := existing
	if dto.Path != nil {
		updated.Path = filepath.Clean(strings.TrimSpace(*dto.Path))
	}
	if dto.Label != nil {
		updated.Label = strings.TrimSpace(*dto.Label)
	}
	if dto.Enabled != nil {
		updated.Enabled = *dto.Enabled
	}

	if err := s.validateWatchFolderPath(updated.Path, id); err != nil {
		return WatchFolderDto{}, err
	}

	var result WatchFolderModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var updateErr error
		result, updateErr = s.Repository.Update(tx, updated)
		return updateErr
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WatchFolderDto{}, ErrWatchFolderNotFound
		}
		return WatchFolderDto{}, fmt.Errorf("UpdateWatchFolder: %w", err)
	}

	log.Println(i18n.Translate("WATCH_FOLDER_UPDATED", result.Path))
	return result.ToDto(), nil
}

func (s *Service) DeleteWatchFolder(id int) error {
	if id <= 0 {
		return ErrInvalidWatchFolderID
	}

	existing, err := s.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWatchFolderNotFound
		}
		return fmt.Errorf("DeleteWatchFolder get by id: %w", err)
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.Delete(tx, id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWatchFolderNotFound
		}
		return fmt.Errorf("DeleteWatchFolder: %w", err)
	}

	log.Println(i18n.Translate("WATCH_FOLDER_DELETED", existing.Path))
	return nil
}

func (s *Service) GetEnabledWatchFolders() ([]WatchFolderModel, error) {
	models, err := s.Repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("GetEnabledWatchFolders: %w", err)
	}

	enabled := make([]WatchFolderModel, 0, len(models))
	for _, model := range models {
		if model.Enabled {
			enabled = append(enabled, model)
		}
	}
	return enabled, nil
}

func (s *Service) UpdateWatchFolderLastScan(id int, lastScanAt time.Time) error {
	if id <= 0 {
		return ErrInvalidWatchFolderID
	}

	err := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdateLastScan(tx, id, lastScanAt)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWatchFolderNotFound
		}
		return fmt.Errorf("UpdateWatchFolderLastScan: %w", err)
	}

	return nil
}

func (s *Service) validateWatchFolderPath(path string, ignoreID int) error {
	if path == "" || path == "." || !filepath.IsAbs(path) {
		return ErrPathNotExists
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ErrPathNotExists
	}
	if err != nil {
		return fmt.Errorf("validateWatchFolderPath stat: %w", err)
	}
	if !info.IsDir() {
		return ErrPathNotExists
	}

	if isSubfolderOfEntryPoint(path) {
		return ErrPathIsSubfolderOfEntryPoint
	}

	all, err := s.Repository.GetAll()
	if err != nil {
		return fmt.Errorf("validateWatchFolderPath get all: %w", err)
	}
	for _, current := range all {
		if current.ID == ignoreID {
			continue
		}
		if filepath.Clean(current.Path) == path {
			return ErrPathAlreadyWatched
		}
	}

	return nil
}

func isSubfolderOfEntryPoint(path string) bool {
	entryPoint := filepath.Clean(strings.TrimSpace(config.AppConfig.EntryPoint))
	if entryPoint == "" || entryPoint == "." {
		return false
	}

	cleanPath := filepath.Clean(path)
	rel, err := filepath.Rel(entryPoint, cleanPath)
	if err != nil {
		return false
	}

	if rel == "." {
		return true
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}
