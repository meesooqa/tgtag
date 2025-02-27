package extensions

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
)

type BaseExtension struct {
	Controllers []controllers.Controller
}

func (e *BaseExtension) ID() string {
	return uuid.New().String()
}

func (e *BaseExtension) GetControllers() []controllers.Controller {
	return e.Controllers
}

func (e *BaseExtension) RegisterRoutes(log *slog.Logger, mux *http.ServeMux, tpl web.Template) {
	if len(e.Controllers) == 0 {
		return
	}
	for _, controller := range e.Controllers {
		controller.Router(log, mux, tpl)
	}
}

func (e *BaseExtension) StaticHandler() (path string, handler http.Handler) {
	// TODO return "/static/", http.FileServer(http.Dir("./templates/default/static"))
	return "/static/", http.FileServer(http.Dir("./templates/default/static"))
}
