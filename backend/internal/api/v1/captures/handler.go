package captures

import (
	"fmt"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    service,
		logService: loggerService,
	}
}

func (h *Handler) UploadCaptureHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "UploadCapture",
		Description: "Uploading a media capture",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	file, err := c.FormFile("file")
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_NO_FILE")})
		return
	}

	dto := CreateCaptureDto{
		Name:      c.PostForm("name"),
		MediaType: c.PostForm("media_type"),
		MimeType:  c.PostForm("mime_type"),
	}

	if dto.Name == "" {
		h.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("name is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_NAME_REQUIRED")})
		return
	}

	if sizeStr := c.PostForm("size"); sizeStr != "" {
		if parsed, parseErr := strconv.ParseInt(sizeStr, 10, 64); parseErr == nil {
			dto.Size = parsed
		}
	}

	loggerModel.SetExtraData(logger.LogExtraData{Data: dto})

	result, err := h.service.UploadCapture(file, dto)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_UPLOAD_FAILED")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetCapturesHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "GetCaptures",
		Description: "Listing captures",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	filter := CaptureFilter{}
	if name := c.Query("name"); name != "" {
		filter.Name = utils.Optional[string]{HasValue: true, Value: name}
	}
	if mediaType := c.Query("media_type"); mediaType != "" {
		filter.MediaType = utils.Optional[string]{HasValue: true, Value: mediaType}
	}

	pagination, err := h.service.GetCaptures(filter, page, pageSize)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_LIST_FAILED")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (h *Handler) GetCaptureByIDHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "GetCaptureByID",
		Description: "Fetching capture by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_ID")})
		return
	}

	capture, err := h.service.GetCaptureByID(id)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_NOT_FOUND")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, capture)
}

func (h *Handler) DeleteCaptureHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "DeleteCapture",
		Description: "Deleting a capture",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_ID")})
		return
	}

	err = h.service.DeleteCapture(id)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_DELETE_FAILED")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
