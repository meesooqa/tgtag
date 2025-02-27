package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/meesooqa/tgtag/internal/config"
	"github.com/meesooqa/tgtag/internal/db"
	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
	"github.com/meesooqa/tgtag/pkg/extensions"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

func main() {
	//loggerFile, cleanup := config.InitLogger("var/log/server.log", slog.LevelDebug)
	//defer cleanup()
	logger := config.InitConsoleLogger(slog.LevelDebug)
	conf, err := config.Load("etc/config.yml")
	if err != nil {
		logger.Error("can't load config", "err", err)
	}

	mongoDB := db.NewMongoDB(logger, conf.Mongo)
	err = mongoDB.Init()
	if err != nil {
		logger.Error("db connection failed", slog.Any("err", err))
	}
	defer mongoDB.Close()

	repo := repositories.NewMessageRepository(logger, mongoDB)
	registerExtensions(repo)

	mux := http.NewServeMux()
	menuData := buildMenuData(extensions.GetAllControllers())
	tpl := web.NewDefaultTemplate(logger, menuData)
	// handle common static
	path, handler := tpl.StaticHandler()
	mux.Handle(path, http.StripPrefix(path, handler))
	// handle extensions
	extensions.RegisterAllRoutes(logger, mux, tpl)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", conf.Server.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	logger.Info("server started", slog.Int("port", conf.Server.Port))
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("http server terminated, %s", err.Error())
	}
}

func buildMenuData(menuControllers []controllers.Controller) []web.MenuItem {
	if len(menuControllers) == 0 {
		return nil
	}
	var result []web.MenuItem
	for _, c := range menuControllers {
		mi := web.MenuItem{Title: c.GetTitle(), Link: c.GetRoute()}
		if len(c.GetChildren()) > 0 {
			for _, cc := range c.GetChildren() {
				si := web.MenuItem{Title: cc.GetTitle(), Link: cc.GetRoute()}
				mi.Children = append(mi.Children, si)
			}
		}
		result = append(result, mi)
	}
	return result
}
