package diary

import (
	"fmt"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(diaryService ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    diaryService,
		logService: loggerService,
	}
}

func (handler *Handler) CreateDiaryHandler(c *gin.Context) {

	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "CreateDiary",
		Description: "Creating a new diary entry",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var diaryDto DiaryDto

	if err := c.ShouldBindJSON(&diaryDto); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	startTime := time.Now()

	diaryDto.StartTime = startTime
	diaryDto.EndTime = utils.Optional[time.Time]{HasValue: false}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: diaryDto,
	})

	diaryResult, err := handler.service.CreateDiary(diaryDto)

	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, diaryResult)
}

func (handler *Handler) DuplicateDiaryHandler(c *gin.Context) {

	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "DuplicateDiary",
		Description: "Duplicating a diary entry",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var diaryId DiaryId

	if err := c.ShouldBindJSON(&diaryId); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: diaryId,
	})
	diaryDto, err := handler.service.DuplicateDiary(diaryId.ID)

	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, diaryDto)
}

func (handler *Handler) GetDiaryHandler(c *gin.Context) {

	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetDiary",
		Description: "Fetching diary entries with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "100"), c)

	filter := DiaryFilter{}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: filter,
	})
	pagination, err := handler.service.GetDiary(filter, page, pageSize)

	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) UpdateDiaryHandler(c *gin.Context) {

	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "UpdateDiary",
		Description: "Updating a diary entry",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	data := c.PostForm("data")
	if data == "" {
		handler.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("data is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}

	diaryDto := DiaryDto{
		Name: data,
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: diaryDto,
	})

	handler.service.UpdateDiary(diaryDto)

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, diaryDto)
}

func (handler *Handler) GetSummaryHandler(c *gin.Context) {

	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetSummary",
		Description: "Fetching diary summary",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	summary, err := handler.service.GetSummary()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, summary)
}
