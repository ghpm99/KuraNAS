package vacuum

import (
	"fmt"
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

func (h *Handler) GetSettingsHandler(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettingsHandler(c *gin.Context) {
	var request UpdateSettingsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	settings, err := h.service.UpdateSettings(request)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) TestConnectionHandler(c *gin.Context) {
	if err := h.service.TestConnection(c.Request.Context()); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DiscoverHandler(c *gin.Context) {
	discovery, err := h.service.Discover(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, discovery)
}

func (h *Handler) GetStatusHandler(c *gin.Context) {
	status, err := h.service.GetStatus(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) SendCommandHandler(c *gin.Context) {
	var request CommandDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	if err := h.service.SendCommand(c.Request.Context(), request.Command); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) SendRemoteControlHandler(c *gin.Context) {
	var request RemoteControlDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	if err := h.service.SendRemoteControl(c.Request.Context(), request.Direction); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetPreferencesHandler(c *gin.Context) {
	preferences, err := h.service.GetPreferences(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (h *Handler) UpdatePreferencesHandler(c *gin.Context) {
	var patch UpdatePreferencesDto
	if err := c.ShouldBindJSON(&patch); err != nil {
		respondWithInvalidBody(c)
		return
	}

	preferences, err := h.service.UpdatePreferences(c.Request.Context(), patch)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (h *Handler) GetConsumablesHandler(c *gin.Context) {
	consumables, err := h.service.GetConsumables(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, consumables)
}

func (h *Handler) ResetConsumableHandler(c *gin.Context) {
	if err := h.service.ResetConsumable(c.Request.Context(), c.Param("part")); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetTotalsHandler(c *gin.Context) {
	totals, err := h.service.GetTotals(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, totals)
}

func (h *Handler) GetCloudSettingsHandler(c *gin.Context) {
	settings, err := h.service.GetCloudSettings()
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateCloudSettingsHandler(c *gin.Context) {
	var request UpdateCloudSettingsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	settings, err := h.service.UpdateCloudSettings(request)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) TestCloudConnectionHandler(c *gin.Context) {
	if err := h.service.TestCloudConnection(c.Request.Context()); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetMapHandler(c *gin.Context) {
	robotMap, err := h.service.GetMap(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, robotMap)
}

func (h *Handler) GetRoomsHandler(c *gin.Context) {
	rooms, err := h.service.GetRooms(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) UpdateRoomHandler(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, fmt.Errorf("%w: %q", ErrRoomNotFound, c.Param("id")))
		return
	}
	var request UpdateRoomDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	room, err := h.service.UpdateRoom(c.Request.Context(), roomID, request)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *Handler) GetPathHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetPath())
}

func (h *Handler) CleanRoomsHandler(c *gin.Context) {
	var request RoomCleaningDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	if err := h.service.CleanRooms(c.Request.Context(), request); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CleanSpotHandler(c *gin.Context) {
	var request SpotCleaningDto
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithInvalidBody(c)
		return
	}

	if err := h.service.CleanSpot(c.Request.Context(), request); err != nil {
		respondWithError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
