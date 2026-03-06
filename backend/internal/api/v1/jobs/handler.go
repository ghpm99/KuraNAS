package jobs

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) GetJobByIDHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_ID_REQUIRED")})
		return
	}

	job, err := h.service.GetJobByID(id)
	if err != nil {
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_JOBS_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_JOBS_FETCH_FAILED")})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (h *Handler) GetJobsHandler(c *gin.Context) {
	page, err := parsePositiveInt(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_INVALID_PAGE")})
		return
	}

	pageSize, err := parsePositiveInt(c.DefaultQuery("page_size", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_INVALID_PAGE_SIZE")})
		return
	}

	filter := JobFilter{}
	utils.GenerateFilterFromContext(c, &filter)

	if rawPriority := strings.TrimSpace(c.Query("priority")); rawPriority != "" {
		if _, parseErr := strconv.Atoi(rawPriority); parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_INVALID_PRIORITY")})
			return
		}
	}

	jobs, err := h.service.ListJobs(filter, page, pageSize)
	if err != nil {
		if errors.Is(err, ErrInvalidPage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_INVALID_PAGE")})
			return
		}
		if errors.Is(err, ErrInvalidPageSize) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_INVALID_PAGE_SIZE")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_JOBS_FETCH_FAILED")})
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) GetJobStepsHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_JOBS_ID_REQUIRED")})
		return
	}

	steps, err := h.service.GetStepsByJobID(id)
	if err != nil {
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_JOBS_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_JOBS_STEPS_FETCH_FAILED")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": steps})
}

func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < 1 {
		return 0, errors.New("must be positive")
	}

	return parsed, nil
}
