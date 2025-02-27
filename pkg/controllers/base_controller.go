package controllers

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/meesooqa/tgtag/internal/web"
)

type BaseController struct {
	Self         ControllerDataProvider
	Log          *slog.Logger
	MethodApi    string
	RouteApi     string
	Method       string
	Route        string
	Title        string
	ContentTpl   string
	Tpl          web.Template
	Children     []Controller
	templates    *template.Template
	fsContentTpl embed.FS
}

func (c *BaseController) Router(log *slog.Logger, mux *http.ServeMux, tpl web.Template, fsContentTpl embed.FS) {
	c.Log = log
	c.Tpl = tpl
	c.fsContentTpl = fsContentTpl
	c.initTemplates()
	// the Children first
	if len(c.GetChildren()) > 0 {
		for _, cc := range c.GetChildren() {
			cc.Router(log, mux, c.Tpl, fsContentTpl)
		}
	}
	// then the parent
	if c.Route != "" {
		mux.HandleFunc(c.Route, c.handlePage)
	}
	if c.RouteApi != "" {
		mux.HandleFunc(c.RouteApi, c.handleApi)
	}
}

func (c *BaseController) GetChildren() []Controller {
	return c.Children
}

func (c *BaseController) AddChildren(cc ...Controller) {
	c.Children = append(c.Children, cc...)
}

func (c *BaseController) GetTitle() string {
	return c.Title
}

func (c *BaseController) GetRoute() string {
	return c.Route
}

func (c *BaseController) handleApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != c.MethodApi {
		c.Log.Error("method is not allowed", slog.String("methodApi", c.MethodApi), slog.String("routeApi", c.RouteApi))
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	data := c.Self.GetApiData(r)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		c.Log.Error("encoding api data", slog.String("methodApi", c.MethodApi), slog.String("routeApi", c.RouteApi), slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *BaseController) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != c.Method {
		c.Log.Error("method is not allowed", slog.String("method", c.Method), slog.String("route", c.Route))
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	data := c.Self.GetTplData(r)
	if err := c.templates.ExecuteTemplate(w, c.Tpl.GetLayoutTpl(), &data); err != nil {
		c.Log.Error("executing template", slog.String("contentTpl", c.ContentTpl), slog.String("route", c.Route), slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *BaseController) initTemplates() {
	tl := c.Tpl.GetTemplatesLocation()

	var files, fsFiles []string
	// layout
	topLevel, err := fs.Glob(os.DirFS(tl), "*.html")
	if err != nil {
		c.Log.Error("finding tpls - topLevel", slog.Any("err", err))
	}
	files = append(files, topLevel...)
	// content
	var subDir, fsSubDir []string
	if c.ContentTpl == "" {
		c.ContentTpl = c.Tpl.GetDefaultContentTpl()
		subDir, err = fs.Glob(os.DirFS(tl), c.ContentTpl)
	} else {
		fsSubDir, err = fs.Glob(c.fsContentTpl, c.ContentTpl)
		fsFiles = append(fsFiles, fsSubDir...)
	}
	if err != nil {
		c.Log.Error("finding tpls - subDir", slog.Any("err", err))
		log.Fatal(err)
	}
	files = append(files, subDir...)
	for i, f := range files {
		files[i] = tl + "/" + f
	}

	c.templates, err = template.ParseFiles(files...)
	if err != nil {
		c.Log.Error("parsing tpls", slog.Any("err", err))
	}
	if len(fsFiles) != 0 {
		for _, fsFile := range fsFiles {
			_, err = c.templates.ParseFS(c.fsContentTpl, fsFile)
			if err != nil {
				c.Log.Error("parsing FS tpls", slog.Any("fsFile", fsFile), slog.Any("err", err))
			}
		}
	}
}
