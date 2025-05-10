package diary

import (
	"fmt"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(diaryService ServiceInterface) *Handler {
	return &Handler{
		service: diaryService,
	}
}

func (handler *Handler) CreateDiaryHandler(c *gin.Context) {
	var diaryDto DiaryDto

	if err := c.ShouldBindJSON(&diaryDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(diaryDto)

	diaryResult, err := handler.service.CreateDiary(diaryDto)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, diaryResult)
}

func (handler *Handler) GetDiaryHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	filter := DiaryFilter{}

	pagination, err := handler.service.GetDiary(filter, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) UpdateDiaryHandler(c *gin.Context) {
	data := c.PostForm("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}
	fmt.Println(data)
	diaryDto := DiaryDto{
		Name: data,
	}
	handler.service.UpdateDiary(diaryDto)
	c.JSON(http.StatusOK, diaryDto)
}

func (handler *Handler) GetSummaryHandler(c *gin.Context) {
	summary, err := handler.service.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
