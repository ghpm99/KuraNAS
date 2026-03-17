package notifications

import (
	"errors"
	"net/http"
	"strconv"

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

func (handler *Handler) ListNotificationsHandler(c *gin.Context) {
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

	filter := NotificationFilter{}

	if notifType := c.Query("type"); notifType != "" {
		filter.Type.Set(notifType)
	}
	if isRead := c.Query("is_read"); isRead != "" {
		if val, err := strconv.ParseBool(isRead); err == nil {
			filter.IsRead.Set(val)
		}
	}

	notifications, err := handler.service.ListNotifications(filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_LIST_NOTIFICATIONS")})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (handler *Handler) GetNotificationByIDHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	notification, err := handler.service.GetNotificationByID(id)
	if err != nil {
		if errors.Is(err, ErrInvalidNotificationID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return
		}
		if errors.Is(err, ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_NOTIFICATION_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_GET_NOTIFICATION")})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func (handler *Handler) GetUnreadCountHandler(c *gin.Context) {
	count, err := handler.service.GetUnreadCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_GET_UNREAD_COUNT")})
		return
	}

	c.JSON(http.StatusOK, count)
}

func (handler *Handler) MarkAsReadHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	err := handler.service.MarkAsRead(id)
	if err != nil {
		if errors.Is(err, ErrInvalidNotificationID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return
		}
		if errors.Is(err, ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_NOTIFICATION_NOT_FOUND")})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_MARK_NOTIFICATION_READ")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (handler *Handler) MarkAllAsReadHandler(c *gin.Context) {
	err := handler.service.MarkAllAsRead()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_MARK_ALL_NOTIFICATIONS_READ")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
