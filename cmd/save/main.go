package main

import (
	"log/slog"
	"sync"

	"github.com/meesooqa/tgtag/internal/config"
	"github.com/meesooqa/tgtag/internal/db"
	"github.com/meesooqa/tgtag/internal/fs"
	"github.com/meesooqa/tgtag/internal/proc"
	"github.com/meesooqa/tgtag/internal/tg"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

func main() {
	logger := config.InitConsoleLogger(slog.LevelDebug)

	var wg sync.WaitGroup
	conf, err := config.Load("etc/config.yml")
	if err != nil {
		logger.Error("can't load config", "err", err)
	}

	wg.Add(1)
	filesChan := make(chan string, 2)
	finder := fs.NewFinder(logger)
	go finder.FindFiles(conf.System.DataPath, filesChan, &wg)

	wg.Add(1)
	mongoDB := db.NewMongoDB(logger, conf.Mongo)
	err = mongoDB.Init()
	if err != nil {
		logger.Error("db connection failed", "err", err)
	}
	defer mongoDB.Close()

	repo := repositories.NewMessageRepository(logger, mongoDB)
	tgService := tg.NewService(logger, conf.System)
	processor := proc.NewProcessor(logger, tgService, repo)
	go processor.ProcessFile(filesChan, &wg)

	wg.Wait()
	logger.Info("all goroutines are done")
}
