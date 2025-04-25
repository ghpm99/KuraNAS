package worker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

func ScanFilesWorker(service files.ServiceInterface) {
	fmt.Println("🔍 Escaneando arquivos...")

	fail := func(path string, err error) error {
		return fmt.Errorf("❌ Erro ao buscar arquivo %s: %v", path, err)
	}

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				i18n.PrintTranslate("ERROR_PERMISSION_DENIED", path)
				return nil
			}
			return fail(path, err)
		}
		name := info.Name()
		fileDto, fileDtoError := service.GetFileByNameAndPath(name, path)

		if fileDtoError != nil {
			if !errors.Is(fileDtoError, sql.ErrNoRows) {
				return fail(path, err)
			}
			i18n.PrintTranslate("FILE_NOT_FOUND_IN_DATABASE", path)
		}

		if err := fileDto.ParseFileInfoToFileDto(info); err != nil {
			return fail(path, err)
		}

		if fileDtoError == nil {
			updated, err := service.UpdateFile(fileDto)
			if err != nil || !updated {
				return fail(path, err)
			}
			i18n.PrintTranslate("FILE_UPDATE_SUCCESS", fileDto.ID)
			return nil
		} else {
			fileDto.Path = path
		}

		fileCreated, err := service.CreateFile(fileDto)

		if err != nil {
			return fail(path, err)
		}
		i18n.PrintTranslate("FILE_CREATE_SUCCESS", fileCreated.ID)
		return nil
	})

	if err != nil {
		log.Printf("❌ Erro ao escanear arquivos: %v", err)
	} else {
		fmt.Println("✅ Escaneamento concluído!")
	}

	// findFilesDeleted(service)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func findFilesDeleted(service files.ServiceInterface) {
	var currentPage = 1
	var pagination, error = service.GetFiles(files.FileFilter{}, currentPage, 20)
	if error != nil {
		log.Printf("❌ Erro ao buscar arquivos: %v", error)
		return
	}
	for {
		for _, file := range pagination.Items {
			if !fileExists(file.Path) {
				fmt.Printf("Arquivo não existe ID: %d, %v\n", file.ID, file.Name)
				file.DeletedAt = utils.Optional[time.Time]{
					HasValue: true,
					Value:    time.Now(),
				}
				_, error := service.UpdateFile(file)
				if error != nil {
					log.Printf("❌ Erro ao deletar arquivo %s: %v\n", file.Path, error)
					continue
				}
			} else {
				fmt.Printf("Arquivo ainda existe ID: %d, %v\n", file.ID, file.Name)
				continue
			}
		}
		if !pagination.Pagination.HasNext {
			break
		}
		currentPage++
		pagination, error = service.GetFiles(files.FileFilter{}, currentPage, 20)
		if error != nil {
			log.Printf("❌ Erro ao buscar arquivos: %v", error)
			break
		}
	}

}
