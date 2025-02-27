package controllers

import (
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/internal/web"
)

type Controller interface {
	Router(log *slog.Logger, mux *http.ServeMux, tpl web.Template)
	GetChildren() []Controller
	AddChildren(cc ...Controller)

	GetRoute() string
	GetTitle() string
}

// ControllerDataProvider is a provider for Controller
type ControllerDataProvider interface {
	GetApiData(r *http.Request) map[string]any
	GetTplData(r *http.Request) map[string]any
}
