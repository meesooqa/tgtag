package main

import (
	"github.com/meesooqa/tgtag/ext/main_ext"
	"github.com/meesooqa/tgtag/pkg/extensions"
	"github.com/meesooqa/tgtag/pkg/repositories"

	"github.com/meesooqa/tgtag-ext-coocc/ext/coocc_ext"
	"github.com/meesooqa/tgtag-ext-dummy/ext/dummy_ext"
)

func registerExtensions(repo repositories.Repository) {
	extensions.Register(main_ext.NewMainExtension(repo))
	extensions.Register(dummy_ext.NewDummyExtension(repo))
	extensions.Register(coocc_ext.NewCooccExtension(repo))
}
