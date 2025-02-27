package extensions

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
)

type BaseExtension struct {
	Name         string
	Controllers  []controllers.Controller
	FsStaticDir  embed.FS
	FsContentTpl embed.FS
}

func (e *BaseExtension) GetName() string {
	return e.Name
}

func (e *BaseExtension) GetControllers() []controllers.Controller {
	return e.Controllers
}

func (e *BaseExtension) RegisterRoutes(log *slog.Logger, mux *http.ServeMux, tpl web.Template) {
	if len(e.Controllers) == 0 {
		return
	}
	for _, controller := range e.Controllers {
		controller.Router(log, mux, tpl, e.FsContentTpl)
	}
}

func (e *BaseExtension) StaticHandler() (string, http.Handler) {
	if _, err := fs.ReadDir(e.FsStaticDir, "static"); err != nil {
		return "", http.NotFoundHandler()
	}
	path := fmt.Sprintf("/static/%s/", e.GetName())
	return path, http.StripPrefix(path, http.FileServer(http.FS(e.FsStaticDir)))
}
