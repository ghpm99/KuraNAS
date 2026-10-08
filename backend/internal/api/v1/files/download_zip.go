package files

import (
	"archive/zip"
	"fmt"
	"io"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxZipDownloadIds       = 1000
	zipDescendantsPageSize  = 500
	zipArchiveContentType   = "application/zip"
	zipDownloadNameTemplate = "kuranas-%s.zip"
	zipDownloadTimeLayout   = "20060102-150405"
)

type zipArchiveWriter struct {
	archive       *zip.Writer
	usedEntryName map[string]bool
}

func (handler *Handler) DownloadZipHandler(c *gin.Context) {
	fileIds, isValid := parseDownloadIds(c)
	if !isValid {
		return
	}

	selectedFiles := make([]FileDto, 0, len(fileIds))
	for _, fileId := range fileIds {
		file, err := handler.service.GetFileById(fileId)
		if err != nil {
			respondFileLookupError(c, err)
			return
		}
		if file.DeletedAt.HasValue {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
			return
		}
		selectedFiles = append(selectedFiles, file)
	}

	archiveName := fmt.Sprintf(zipDownloadNameTemplate, time.Now().Format(zipDownloadTimeLayout))
	writer := beginZipResponse(c, archiveName)
	defer writer.archive.Close()

	for _, file := range selectedFiles {
		if file.Type == Directory {
			handler.addFolderToZip(writer, file, file.Name+"/")
			continue
		}
		handler.addFileToZip(writer, file, file.Name)
	}
}

func (handler *Handler) streamFolderAsZip(c *gin.Context, folder FileDto) {
	writer := beginZipResponse(c, folder.Name+".zip")
	defer writer.archive.Close()

	handler.addFolderToZip(writer, folder, "")
}

func beginZipResponse(c *gin.Context, archiveName string) *zipArchiveWriter {
	c.Header("Content-Type", zipArchiveContentType)
	c.Header("Content-Disposition", buildAttachmentDisposition(archiveName))
	c.Status(http.StatusOK)
	return &zipArchiveWriter{
		archive:       zip.NewWriter(c.Writer),
		usedEntryName: map[string]bool{},
	}
}

func parseDownloadIds(c *gin.Context) ([]int, bool) {
	rawIds := make([]string, 0)
	for _, queryValue := range c.QueryArray("ids") {
		rawIds = append(rawIds, strings.Split(queryValue, ",")...)
	}

	seenIds := map[int]bool{}
	fileIds := make([]int, 0, len(rawIds))
	for _, rawId := range rawIds {
		rawId = strings.TrimSpace(rawId)
		if rawId == "" {
			continue
		}
		fileId, err := strconv.Atoi(rawId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return nil, false
		}
		if seenIds[fileId] {
			continue
		}
		seenIds[fileId] = true
		fileIds = append(fileIds, fileId)
	}

	if len(fileIds) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return nil, false
	}
	if len(fileIds) > maxZipDownloadIds {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.Translate("ERROR_DOWNLOAD_TOO_MANY_ITEMS", maxZipDownloadIds)})
		return nil, false
	}
	return fileIds, true
}

func (handler *Handler) addFolderToZip(writer *zipArchiveWriter, folder FileDto, entryPrefix string) {
	folderPrefix := folder.Path + string(filepath.Separator)

	for page := 1; ; page++ {
		descendants, err := handler.service.GetFilesByPathPrefix(folderPrefix, page, zipDescendantsPageSize)
		if err != nil {
			applog.ErrorWithStack("files: zip folder listing failed", err, "folder", folder.Path)
			return
		}

		for _, descendant := range descendants.Items {
			if descendant.Type != File || descendant.DeletedAt.HasValue || !strings.HasPrefix(descendant.Path, folderPrefix) {
				continue
			}
			relativePath := filepath.ToSlash(strings.TrimPrefix(descendant.Path, folderPrefix))
			handler.addFileToZip(writer, descendant, entryPrefix+relativePath)
		}

		if !descendants.Pagination.HasNext {
			return
		}
	}
}

func (handler *Handler) addFileToZip(writer *zipArchiveWriter, file FileDto, entryName string) {
	content, err := os.Open(file.ResolveContentPath())
	if err != nil {
		applog.Warn("files: zip entry skipped, cannot open content", "path", file.Path, "error", err.Error())
		return
	}
	defer content.Close()

	contentInfo, err := content.Stat()
	if err != nil || !contentInfo.Mode().IsRegular() {
		applog.Warn("files: zip entry skipped, not a readable regular file", "path", file.Path)
		return
	}

	entryWriter, err := writer.archive.CreateHeader(&zip.FileHeader{
		Name:     writer.uniqueEntryName(entryName),
		Method:   zip.Store,
		Modified: contentInfo.ModTime(),
	})
	if err != nil {
		applog.Warn("files: zip entry skipped, cannot create header", "path", file.Path, "error", err.Error())
		return
	}

	if _, err := io.Copy(entryWriter, content); err != nil {
		applog.Warn("files: zip entry truncated, read failed", "path", file.Path, "error", err.Error())
	}
}

func (writer *zipArchiveWriter) uniqueEntryName(entryName string) string {
	if !writer.usedEntryName[entryName] {
		writer.usedEntryName[entryName] = true
		return entryName
	}

	extension := filepath.Ext(entryName)
	baseName := strings.TrimSuffix(entryName, extension)
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s (%d)%s", baseName, suffix, extension)
		if !writer.usedEntryName[candidate] {
			writer.usedEntryName[candidate] = true
			return candidate
		}
	}
}
