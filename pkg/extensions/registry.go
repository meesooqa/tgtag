package extensions

import (
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
)

var modules []Extension

func Register(module Extension) {
	modules = append(modules, module)
}

func RegisterAllRoutes(log *slog.Logger, mux *http.ServeMux, tpl web.Template) {
	for _, module := range modules {
		module.RegisterRoutes(log, mux, tpl)
		path, handler := module.StaticHandler()
		if path != "" {
			mux.Handle(path, http.StripPrefix(path, handler))
		}
	}
}

func GetAllControllers() []controllers.Controller {
	list := make([]controllers.Controller, 0)
	for _, module := range modules {
		moduleControllers := module.GetControllers()
		if len(moduleControllers) > 0 {
			for _, controller := range moduleControllers {
				list = append(list, controller)
			}
		}
	}
	return list
}
