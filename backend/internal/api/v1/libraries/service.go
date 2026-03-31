package libraries

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

func (s *Service) GetLibraries() ([]LibraryDto, error) {
	if err := s.ResolveLibraries(); err != nil {
		return nil, fmt.Errorf("GetLibraries resolve: %w", err)
	}

	models, err := s.Repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("GetLibraries: %w", err)
	}

	modelByCategory := make(map[LibraryCategory]LibraryModel, len(models))
	for _, model := range models {
		modelByCategory[model.Category] = model
	}

	dtos := make([]LibraryDto, 0, len(AllCategories))
	for _, category := range AllCategories {
		model, exists := modelByCategory[category]
		if !exists {
			continue
		}
		dtos = append(dtos, model.ToDto())
	}

	return dtos, nil
}

func (s *Service) GetLibraryByCategory(category LibraryCategory) (LibraryDto, error) {
	if !category.IsValid() {
		return LibraryDto{}, ErrInvalidCategory
	}

	model, err := s.Repository.GetByCategory(category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			entryPoint := filepath.Clean(config.AppConfig.EntryPoint)
			resolvedPath := resolveLibraryPath(entryPoint, category)
			resolveErr := s.withTransaction(func(tx *sql.Tx) error {
				var upsertErr error
				model, upsertErr = s.Repository.Upsert(tx, LibraryModel{
					Category: category,
					Path:     resolvedPath,
				})
				return upsertErr
			})
			if resolveErr != nil {
				return LibraryDto{}, fmt.Errorf("GetLibraryByCategory upsert resolve: %w", resolveErr)
			}
			return model.ToDto(), nil
		}
		return LibraryDto{}, fmt.Errorf("GetLibraryByCategory: %w", err)
	}

	return model.ToDto(), nil
}

func (s *Service) UpdateLibrary(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error) {
	if !category.IsValid() {
		return LibraryDto{}, ErrInvalidCategory
	}

	cleanPath := filepath.Clean(dto.Path)

	if err := validatePathIsSubfolder(cleanPath); err != nil {
		return LibraryDto{}, err
	}

	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return LibraryDto{}, ErrPathNotExists
	}
	if err != nil {
		return LibraryDto{}, fmt.Errorf("UpdateLibrary stat path: %w", err)
	}
	if !info.IsDir() {
		return LibraryDto{}, ErrPathNotExists
	}

	var result LibraryModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var upsertErr error
		result, upsertErr = s.Repository.Upsert(tx, LibraryModel{
			Category: category,
			Path:     cleanPath,
		})
		return upsertErr
	})

	if err != nil {
		return LibraryDto{}, fmt.Errorf("UpdateLibrary: %w", err)
	}

	log.Println(i18n.Translate("LIBRARY_UPDATE_SUCCESS", string(category), cleanPath))
	return result.ToDto(), nil
}

func (s *Service) ResolveLibraries() error {
	if strings.TrimSpace(config.AppConfig.EntryPoint) == "" {
		return fmt.Errorf("ResolveLibraries: ENTRY_POINT is not configured")
	}

	entryPoint := filepath.Clean(config.AppConfig.EntryPoint)
	if entryPoint == "" || entryPoint == "." {
		return fmt.Errorf("ResolveLibraries: ENTRY_POINT is not configured")
	}

	for _, category := range AllCategories {
		if err := s.resolveCategory(entryPoint, category); err != nil {
			log.Printf("ResolveLibraries: failed to resolve %s: %v", category, err)
		}
	}

	return nil
}

func (s *Service) resolveCategory(entryPoint string, category LibraryCategory) error {
	existing, err := s.Repository.GetByCategory(category)
	if err == nil && existing.Path != "" {
		if _, statErr := os.Stat(existing.Path); statErr == nil {
			log.Println(i18n.Translate("LIBRARY_RESOLVED", string(category), existing.Path))
			return nil
		}
	}

	resolvedPath := resolveLibraryPath(entryPoint, category)

	return s.withTransaction(func(tx *sql.Tx) error {
		result, upsertErr := s.Repository.Upsert(tx, LibraryModel{
			Category: category,
			Path:     resolvedPath,
		})
		if upsertErr != nil {
			return upsertErr
		}
		log.Println(i18n.Translate("LIBRARY_RESOLVED", string(category), result.Path))
		return nil
	})
}

func validatePathIsSubfolder(path string) error {
	entryPoint := filepath.Clean(config.AppConfig.EntryPoint)
	cleanPath := filepath.Clean(path)
	relative, err := filepath.Rel(entryPoint, cleanPath)
	if err != nil {
		return ErrPathNotSubfolder
	}
	if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return ErrPathNotSubfolder
	}

	return nil
}
