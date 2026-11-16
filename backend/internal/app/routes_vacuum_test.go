package app

import (
	"net/http"
	"testing"

	"nas-go/api/internal/api/v1/vacuum"

	"github.com/gin-gonic/gin"
)

func TestRegisterVacuumRoutesRegistersTheContractedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	context := &AppContext{Vacuum: &VacuumContext{Handler: vacuum.NewHandler(nil)}}

	RegisterVacuumRoutes(router.Group("/api/v1"), context)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}
	wanted := []string{
		http.MethodGet + " /api/v1/vacuum/settings",
		http.MethodPut + " /api/v1/vacuum/settings",
		http.MethodPost + " /api/v1/vacuum/connection-test",
		http.MethodPost + " /api/v1/vacuum/discovery",
		http.MethodGet + " /api/v1/vacuum/status",
		http.MethodPost + " /api/v1/vacuum/commands",
		http.MethodPost + " /api/v1/vacuum/remote-control",
		http.MethodGet + " /api/v1/vacuum/preferences",
		http.MethodPatch + " /api/v1/vacuum/preferences",
		http.MethodGet + " /api/v1/vacuum/consumables",
		http.MethodPost + " /api/v1/vacuum/consumables/:part/reset",
		http.MethodGet + " /api/v1/vacuum/totals",
	}
	for _, route := range wanted {
		if !registered[route] {
			t.Errorf("route not registered: %s", route)
		}
	}
}

func TestRegisterVacuumRoutesIgnoresMissingContext(t *testing.T) {
	router := gin.New()

	RegisterVacuumRoutes(router.Group("/api/v1"), nil)
	RegisterVacuumRoutes(router.Group("/api/v1"), &AppContext{})

	if len(router.Routes()) != 0 {
		t.Fatalf("no route may be registered without a handler")
	}
}
