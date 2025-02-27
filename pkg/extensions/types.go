package extensions

import (
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
)

// Extension describes extension features
type Extension interface {
	// GetName returns Name of extension
	GetName() string

	// GetControllers returns Controllers of extension
	GetControllers() []controllers.Controller

	// RegisterRoutes adds new API-route
	RegisterRoutes(log *slog.Logger, mux *http.ServeMux, tpl web.Template)

	// StaticHandler returns http.Handler, handler of extension's static files
	StaticHandler() (string, http.Handler)
}
