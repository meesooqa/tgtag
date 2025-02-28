package main_ext

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/data"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type GroupController struct {
	controllers.BaseController
	provider data.Provider
}

func NewGroupController(repo repositories.Repository) *GroupController {
	c := &GroupController{
		BaseController: controllers.BaseController{
			MethodApi: http.MethodGet,
			RouteApi:  "/api/groups",
		},
		provider: NewGroupDataProvider(repo),
	}
	c.Self = c
	return c
}

func (c *GroupController) GetApiData(r *http.Request) map[string]any {
	c.provider.SetLogger(c.Log)
	apiData, err := c.provider.GetData(context.TODO(), r.URL.Query().Get("group"))
	if err != nil {
		c.Log.Error("getting api data", slog.Any("err", err))
		return nil
	}
	return map[string]any{"data": apiData}
}

func (c *GroupController) GetTplData(r *http.Request) map[string]any {
	return nil
}
