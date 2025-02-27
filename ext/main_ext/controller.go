package main_ext

import (
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type MainController struct {
	controllers.BaseController
}

func NewMainController(repo repositories.Repository) *MainController {
	c := &MainController{controllers.BaseController{
		MethodApi:  http.MethodGet,
		RouteApi:   "/api/main",
		Method:     http.MethodGet,
		Route:      "/",
		Title:      "Main Extension Page",
		ContentTpl: "template/content/home.html",
		Children: []controllers.Controller{
			NewGroupController(repo),
		},
	}}
	c.Self = c
	return c
}

func (c *MainController) GetApiData(r *http.Request) map[string]any {
	return map[string]any{
		"message": "Hello from MainExtension!",
	}
}

func (c *MainController) GetTplData(r *http.Request) map[string]any {
	data, err := c.Tpl.GetData(r, map[string]any{
		"Title":    c.GetTitle(),
		"Message":  "Привет от шаблона с использованием Go!",
		"IndexVar": "IndexVar value",
	})
	if err != nil {
		c.Log.Error("getting tpl data", slog.Any("err", err))
		return nil
	}
	return data
}
