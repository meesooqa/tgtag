package main_ext

import (
	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/extensions"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type MainExtension struct {
	extensions.BaseExtension
}

func NewMainExtension(repo repositories.Repository) *MainExtension {
	return &MainExtension{extensions.BaseExtension{
		Controllers: []controllers.Controller{
			NewMainController(repo),
		},
	}}
}
