package storageroots

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
)

type Service struct {
	Repository   RepositoryInterface
	IndexTrigger IndexTrigger
}

func NewService(repository RepositoryInterface, indexTrigger IndexTrigger) *Service {
	return &Service{Repository: repository, IndexTrigger: indexTrigger}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

// ReloadRegistry loads the table into the in-memory registry. With an empty
// table and a configured ENTRY_POINT it seeds the legacy root first, so
// existing installs migrate without user action.
func (s *Service) ReloadRegistry() error {
	models, err := s.Repository.GetAll()
	if err != nil {
		return err
	}

	if len(models) == 0 {
		seeded, seedErr := s.seedFromEntryPoint()
		if seedErr != nil {
			return seedErr
		}
		models = seeded
	}

	registry := make([]roots.Root, 0, len(models))
	for index := range models {
		registry = append(registry, models[index].toRegistryRoot())
	}
	roots.Set(registry)
	return nil
}

func (s *Service) seedFromEntryPoint() ([]StorageRootModel, error) {
	entryPoint := strings.TrimSpace(config.AppConfig.EntryPoint)
	if entryPoint == "" {
		return nil, nil
	}
	cleanPath := filepath.Clean(entryPoint)

	var created StorageRootModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		created, createErr = s.Repository.Create(tx, StorageRootModel{
			Path:    cleanPath,
			Label:   filepath.Base(cleanPath),
			Enabled: true,
		})
		return createErr
	})
	if err != nil {
		return nil, fmt.Errorf("seed storage root from ENTRY_POINT: %w", err)
	}

	log.Printf("storageroots: seeded %q as the primary storage root", cleanPath)
	return []StorageRootModel{created}, nil
}

func (s *Service) GetRoots() ([]StorageRootDto, error) {
	models, err := s.Repository.GetAll()
	if err != nil {
		return nil, err
	}

	dtos := make([]StorageRootDto, 0, len(models))
	for index := range models {
		dtos = append(dtos, models[index].ToDto())
	}
	return dtos, nil
}

func validateRootPath(candidate string, registered []StorageRootModel) (string, error) {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" || !filepath.IsAbs(trimmed) {
		return "", ErrInvalidRootPath
	}
	cleanPath := filepath.Clean(trimmed)

	info, err := os.Stat(cleanPath)
	if err != nil || !info.IsDir() {
		return "", ErrInvalidRootPath
	}

	separator := string(filepath.Separator)
	for _, existing := range registered {
		if existing.Path == cleanPath {
			return "", ErrDuplicateRoot
		}
		if strings.HasPrefix(cleanPath, existing.Path+separator) ||
			strings.HasPrefix(existing.Path, cleanPath+separator) {
			return "", ErrOverlappingRoot
		}
	}

	return cleanPath, nil
}

func validateRootLabel(candidate string, fallback string, registered []StorageRootModel, selfID int) (string, error) {
	label := strings.TrimSpace(candidate)
	if label == "" {
		label = fallback
	}
	if label == "" || strings.ContainsAny(label, "/\\") {
		return "", ErrInvalidRootLabel
	}
	for _, existing := range registered {
		if existing.ID != selfID && existing.Label == label {
			return "", ErrDuplicateRoot
		}
	}
	return label, nil
}

func (s *Service) CreateRoot(request CreateStorageRootDto) (StorageRootDto, error) {
	registered, err := s.Repository.GetAll()
	if err != nil {
		return StorageRootDto{}, err
	}

	cleanPath, err := validateRootPath(request.Path, registered)
	if err != nil {
		return StorageRootDto{}, err
	}
	label, err := validateRootLabel(request.Label, filepath.Base(cleanPath), registered, 0)
	if err != nil {
		return StorageRootDto{}, err
	}

	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}

	var created StorageRootModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		created, createErr = s.Repository.Create(tx, StorageRootModel{
			Path:    cleanPath,
			Label:   label,
			Enabled: enabled,
		})
		return createErr
	})
	if err != nil {
		return StorageRootDto{}, err
	}

	if reloadErr := s.ReloadRegistry(); reloadErr != nil {
		log.Printf("storageroots: registry reload after create failed: %v", reloadErr)
	}

	// A new enabled root must enter the index without waiting for a reboot.
	if enabled && s.IndexTrigger != nil {
		s.IndexTrigger.ScanDirTask(created.Path)
	}

	return created.ToDto(), nil
}

// isPrimary reports whether id is the first registered root — the anchor of
// the legacy bare relative paths, which must always stay enabled.
func isPrimary(id int, registered []StorageRootModel) bool {
	return len(registered) > 0 && registered[0].ID == id
}

func (s *Service) UpdateRoot(id int, request UpdateStorageRootDto) (StorageRootDto, error) {
	registered, err := s.Repository.GetAll()
	if err != nil {
		return StorageRootDto{}, err
	}

	current, found, err := s.Repository.GetByID(id)
	if err != nil {
		return StorageRootDto{}, err
	}
	if !found {
		return StorageRootDto{}, ErrRootNotFound
	}

	label, err := validateRootLabel(request.Label, current.Label, registered, id)
	if err != nil {
		return StorageRootDto{}, err
	}

	enabled := current.Enabled
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	if !enabled && isPrimary(id, registered) {
		return StorageRootDto{}, ErrPrimaryRootImmutable
	}

	var updated StorageRootModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var updateErr error
		updated, updateErr = s.Repository.Update(tx, StorageRootModel{
			ID:      id,
			Label:   label,
			Enabled: enabled,
		})
		return updateErr
	})
	if err != nil {
		return StorageRootDto{}, err
	}

	if reloadErr := s.ReloadRegistry(); reloadErr != nil {
		log.Printf("storageroots: registry reload after update failed: %v", reloadErr)
	}

	return updated.ToDto(), nil
}

func (s *Service) DeleteRoot(id int) error {
	registered, err := s.Repository.GetAll()
	if err != nil {
		return err
	}
	if isPrimary(id, registered) {
		return ErrPrimaryRootImmutable
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.Delete(tx, id)
	})
	if err != nil {
		return err
	}

	if reloadErr := s.ReloadRegistry(); reloadErr != nil {
		log.Printf("storageroots: registry reload after delete failed: %v", reloadErr)
	}

	return nil
}
