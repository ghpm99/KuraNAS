package diary

import (
	"fmt"
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
	data := c.PostForm("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}
	fmt.Println(data)
	diaryDto := DiaryDto{
		Name: data,
	}
	handler.service.CreateDiary(diaryDto)
}
