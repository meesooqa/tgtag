package main_ext

import (
	"embed"

	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/extensions"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

//go:embed template/content/*.html
var fsContentTpl embed.FS

//go:embed template/static
var fsStaticDir embed.FS

type MainExtension struct {
	extensions.BaseExtension
}

func NewMainExtension(repo repositories.Repository) *MainExtension {
	return &MainExtension{extensions.BaseExtension{
		Name:         "main_ext",
		FsContentTpl: fsContentTpl,
		FsStaticDir:  fsStaticDir,
		Controllers: []controllers.Controller{
			NewMainController(repo),
		},
	}}
}
