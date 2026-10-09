package assistant

import (
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) ChatHandler(c *gin.Context) {
	if handler.service == nil {
		handler.serviceUnavailable(c)
		return
	}

	var req ChatRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	response, err := handler.service.Chat(req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChatStreamHandler answers over Server-Sent Events: one `delta` event per
// content chunk, then a final `done` event with the full message (or an `error`
// event if generation fails after the stream has started).
func (handler *Handler) ChatStreamHandler(c *gin.Context) {
	if handler.service == nil {
		handler.serviceUnavailable(c)
		return
	}

	var req ChatRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		applog.Error("assistant: response writer is not a flusher", "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	onDelta := func(delta string) error {
		writeSSE(c.Writer, "delta", StreamDeltaDto{Content: delta})
		flusher.Flush()
		return nil
	}

	response, err := handler.service.ChatStream(req, onDelta)
	if err != nil {
		applog.ErrorWithStack("assistant: chat stream failed", err, "ip", c.ClientIP())
		writeSSE(c.Writer, "error", StreamErrorDto{Error: i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		flusher.Flush()
		return
	}

	writeSSE(c.Writer, "done", response)
	flusher.Flush()
}

func (handler *Handler) ListConversationsHandler(c *gin.Context) {
	if handler.service == nil {
		handler.serviceUnavailable(c)
		return
	}

	conversations, err := handler.service.ListConversations()
	if err != nil {
		applog.ErrorWithStack("assistant: list conversations failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func (handler *Handler) GetMessagesHandler(c *gin.Context) {
	if handler.service == nil {
		handler.serviceUnavailable(c)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	messages, err := handler.service.GetMessages(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (handler *Handler) DeleteConversationHandler(c *gin.Context) {
	if handler.service == nil {
		handler.serviceUnavailable(c)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.DeleteConversation(id); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// serviceUnavailable answers 500 when the assistant has no AI service wired
// (config-time issue). It is logged because a silently 500-ing assistant is
// otherwise invisible in the forensic log.
func (handler *Handler) serviceUnavailable(c *gin.Context) {
	applog.Error("assistant: service unavailable (AI not configured)", "path", c.FullPath(), "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidConversation):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
	case errors.Is(err, ErrConversationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
	case errors.Is(err, ErrAIUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
	default:
		applog.ErrorWithStack("assistant: service error", err, "path", c.FullPath(), "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
	}
}

func writeSSE(w http.ResponseWriter, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}
