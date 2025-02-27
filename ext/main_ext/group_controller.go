package main_ext

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type GroupController struct {
	controllers.BaseController
	repo repositories.Repository
}

func NewGroupController(repo repositories.Repository) *GroupController {
	c := &GroupController{
		BaseController: controllers.BaseController{
			MethodApi: http.MethodGet,
			RouteApi:  "/api/groups",
		},
		repo: repo,
	}
	c.Self = c
	return c
}

func (c *GroupController) GetApiData(r *http.Request) map[string]any {
	data, err := c.repo.GetGroups(context.TODO())
	if err != nil {
		c.Log.Error("getting api data", slog.Any("err", err))
		return nil
	}
	return map[string]any{"data": data}
}

func (c *GroupController) GetTplData(r *http.Request) map[string]any {
	return nil
}
