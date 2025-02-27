package main

import (
	"github.com/meesooqa/tgtag/ext/main_ext"
	"github.com/meesooqa/tgtag/pkg/extensions"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

func registerExtensions(repo repositories.Repository) {
	extensions.Register(main_ext.NewMainExtension(repo))
}
