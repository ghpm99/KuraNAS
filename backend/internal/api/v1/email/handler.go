package email

import (
	"errors"
	"fmt"
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

func (h *Handler) GetAccountsHandler(c *gin.Context) {
	accounts, err := h.service.ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_EMAIL_ACCOUNTS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) DeleteAccountHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := h.service.DeleteAccount(id); err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_NOT_FOUND")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_REMOVE_FAILED")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("EMAIL_ACCOUNT_REMOVED")})
}

func (h *Handler) UpdateSyncEnabledHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var request UpdateSyncEnabledDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := h.service.SetSyncEnabled(id, request.SyncEnabled); err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_NOT_FOUND")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_UPDATE_FAILED")})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GoogleAuthURLHandler(c *gin.Context) {
	dto, err := h.service.GoogleAuthURL()
	if err != nil {
		if errors.Is(err, ErrProviderNotConfigured) {
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("EMAIL_OAUTH_NOT_CONFIGURED")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_LINK_FAILED")})
		return
	}
	c.JSON(http.StatusOK, dto)
}

// GoogleCallbackHandler lands in the user's browser at the end of the consent
// flow, so it answers a minimal HTML page instead of JSON.
func (h *Handler) GoogleCallbackHandler(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if errCode := c.Query("error"); errCode != "" || code == "" {
		h.renderCallbackPage(c, http.StatusBadRequest, i18n.GetMessage("EMAIL_ACCOUNT_LINK_FAILED"))
		return
	}

	if err := h.service.HandleGoogleCallback(state, code); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidOAuthState) {
			status = http.StatusBadRequest
		}
		h.renderCallbackPage(c, status, i18n.GetMessage("EMAIL_ACCOUNT_LINK_FAILED"))
		return
	}

	h.renderCallbackPage(c, http.StatusOK, i18n.GetMessage("EMAIL_ACCOUNT_LINKED"))
}

func (h *Handler) renderCallbackPage(c *gin.Context, status int, message string) {
	page := fmt.Sprintf("<!doctype html><html><body><p>%s</p></body></html>", message)
	c.Data(status, "text/html; charset=utf-8", []byte(page))
}

func (h *Handler) MicrosoftDeviceCodeHandler(c *gin.Context) {
	dto, err := h.service.StartMicrosoftDeviceCode()
	if err != nil {
		if errors.Is(err, ErrProviderNotConfigured) {
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("EMAIL_OAUTH_NOT_CONFIGURED")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_LINK_FAILED")})
		return
	}

	dto.Message = i18n.Translate("EMAIL_OAUTH_DEVICE_CODE_PROMPT", dto.VerificationURI, dto.UserCode)
	c.JSON(http.StatusOK, dto)
}

func (h *Handler) MicrosoftDeviceCodeStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.MicrosoftDeviceCodeStatus())
}

func (h *Handler) GetMessagesHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	if c.IsAborted() {
		return
	}

	messages, err := h.service.ListMessages(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_EMAIL_MESSAGES_LOAD")})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *Handler) SyncAccountHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	jobID, err := h.service.EnqueueSync(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("EMAIL_ACCOUNT_NOT_FOUND")})
		case errors.Is(err, ErrSyncUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("EMAIL_SYNC_UNAVAILABLE")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_SYNC_ENQUEUE_FAILED")})
		}
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "message": i18n.GetMessage("EMAIL_SYNC_ENQUEUED")})
}

func (h *Handler) GetMessageSummaryHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	analysis, err := h.service.GetMessageAnalysis(id)
	if err != nil {
		if errors.Is(err, ErrAnalysisNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("EMAIL_ANALYSIS_UNAVAILABLE")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_EMAIL_MESSAGES_LOAD")})
		return
	}
	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) GetProviderHandler(c *gin.Context) {
	dto, err := h.service.GetProviderPreference()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_PROVIDER_LOAD_FAILED")})
		return
	}
	c.JSON(http.StatusOK, dto)
}

func (h *Handler) SetProviderHandler(c *gin.Context) {
	var request ProviderPreferenceDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	dto, err := h.service.SetProviderPreference(request.Provider)
	if err != nil {
		if errors.Is(err, ErrInvalidProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("EMAIL_PROVIDER_INVALID")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("EMAIL_PROVIDER_UPDATE_FAILED")})
		return
	}
	c.JSON(http.StatusOK, dto)
}

// DisabledHandler answers every e-mail route when EMAIL_TOKEN_KEY is not
// configured: the feature refuses to turn on and nothing is ever stored.
func DisabledHandler(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("EMAIL_FEATURE_DISABLED_NO_KEY")})
}
