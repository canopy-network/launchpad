package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RouteInfo represents information about a registered route
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Middlewares int    `json:"middleware_count"`
}

// ListRoutes returns an HTTP handler that lists all registered routes
func ListRoutes(router *chi.Mux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var routes []RouteInfo

		// Walk through all registered routes
		chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			routes = append(routes, RouteInfo{
				Method:      method,
				Path:        route,
				Middlewares: len(middlewares),
			})
			return nil
		})

		// Return routes as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":  routes,
			"count": len(routes),
		})
	}
}
