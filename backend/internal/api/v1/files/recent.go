package files

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
)

func (handler *Handler) GetRecentFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetRecentFiles",
		Description: "Fetching recent files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	recentFiles, err := handler.recentFileService.GetRecentFiles(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, recentFiles)
}

func (handler *Handler) GetRecentAccessByFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetRecentAccessByFile",
		Description: "Fetching recent access by file ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	recentFiles, err := handler.recentFileService.GetRecentAccessByFileID(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, recentFiles)
}

func (handler *Handler) StarreFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StarFile",
		Description: "Starring a file by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	file.Starred = !file.Starred

	result, err := handler.service.UpdateFile(file)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"success": result})
}

type RecentFileService struct {
	Repo RecentFileRepositoryInterface
}

func NewRecentFileService(repo RecentFileRepositoryInterface) *RecentFileService {
	return &RecentFileService{Repo: repo}
}

var DEFAULT_LIMIT = 10

func (s *RecentFileService) RegisterAccess(ip string, fileID int) error {
	return s.Repo.Upsert(ip, fileID)
}

func (s *RecentFileService) GetRecentFiles(page int, limit int) ([]RecentFileDto, error) {
	if limit <= 0 {
		limit = DEFAULT_LIMIT
	}

	if page < 1 {
		page = 1
	}
	recentAccess, err := s.Repo.GetRecentFiles(page, limit)
	if err != nil {
		return nil, err
	}
	var result []RecentFileDto
	for _, access := range recentAccess {
		dto := access.ToDto()
		result = append(result, dto)
	}
	return result, nil
}

func (s *RecentFileService) DeleteRecentFile(ip string, fileID int) error {
	return s.Repo.Delete(ip, fileID)
}

func (s *RecentFileService) GetRecentAccessByFileID(fileID int) ([]RecentFileDto, error) {
	recentAccess, err := s.Repo.GetByFileID(fileID)
	if err != nil {
		return nil, err
	}
	var result []RecentFileDto
	for _, access := range recentAccess {
		dto := access.ToDto()
		result = append(result, dto)
	}
	return result, nil
}

type RecentFileRepository struct {
	DbContext *database.DbContext
}

func NewRecentFileRepository(db *database.DbContext) *RecentFileRepository {
	return &RecentFileRepository{DbContext: db}
}

func (r *RecentFileRepository) Upsert(ip string, fileID int) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.UpsertRecentFileQuery,
			ip, fileID,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("falha ao realizar upsert de arquivo recente: %w", err)
	}
	return nil
}

func (r *RecentFileRepository) GetRecentFiles(page int, pageSize int) ([]RecentFileModel, error) {
	var result []RecentFileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetRecentFilesQuery,
			pageSize+1,
			utils.CalculateOffset(page, pageSize),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rf RecentFileModel
			if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
				return err
			}
			result = append(result, rf)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao obter arquivos recentes: %w", err)
	}
	return result, nil
}

func (r *RecentFileRepository) Delete(ip string, fileID int) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.DeleteRecentFileQuery,
			ip, fileID,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("falha ao deletar arquivo recente: %w", err)
	}
	return nil
}

func (r *RecentFileRepository) GetByFileID(fileID int) ([]RecentFileModel, error) {
	var result []RecentFileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetRecentByFileIDQuery,
			fileID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rf RecentFileModel
			if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
				return err
			}
			result = append(result, rf)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao obter arquivos recentes por ID do arquivo: %w", err)
	}
	return result, nil
}
