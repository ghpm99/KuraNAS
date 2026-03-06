package jobs

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) GetJobByIDHandler(c *gin.Context) {
	jobID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
		return
	}

	job, err := handler.service.GetJobByID(jobID)
	if err != nil {
		if errors.Is(err, ErrInvalidJobID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
			return
		}
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_JOB_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_GET_JOB")})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (handler *Handler) ListJobsHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "20"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	filter := JobFilter{}

	if status := c.Query("status"); status != "" {
		filter.Status.Set(status)
	}
	if jobType := c.Query("type"); jobType != "" {
		filter.Type.Set(jobType)
	}
	if priority := c.Query("priority"); priority != "" {
		filter.Priority.Set(priority)
	}

	jobs, err := handler.service.ListJobs(filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_LIST_JOBS")})
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (handler *Handler) GetStepsByJobIDHandler(c *gin.Context) {
	jobID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
		return
	}

	steps, err := handler.service.GetStepsByJobID(jobID)
	if err != nil {
		if errors.Is(err, ErrInvalidJobID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
			return
		}
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_JOB_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_GET_JOB_STEPS")})
		return
	}

	c.JSON(http.StatusOK, steps)
}

func (handler *Handler) CancelJobHandler(c *gin.Context) {
	jobID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
		return
	}

	err := handler.service.CancelJob(jobID)
	if err != nil {
		if errors.Is(err, ErrInvalidJobID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOB_INVALID_ID")})
			return
		}
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_JOB_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_JOB_CANCEL")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_JOB_CANCEL_SUCCESS")})
}
