package proc

import (
	"log/slog"
	"sync"

	"github.com/meesooqa/tgtag/internal/tg"
	"github.com/meesooqa/tgtag/pkg/models"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type Processor struct {
	log     *slog.Logger
	service tg.Service
	repo    repositories.Repository
}

func NewProcessor(log *slog.Logger, service tg.Service, repo repositories.Repository) *Processor {
	return &Processor{
		log:     log,
		service: service,
		repo:    repo,
	}
}

func (p *Processor) ProcessFile(filesChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	messagesChan := make(chan models.Message, 10)

	var wgm sync.WaitGroup
	wgm.Add(1)
	go func() {
		defer wgm.Done()
		p.repo.UpsertMany(messagesChan)
	}()

	for filename := range filesChan {
		if err := p.service.ParseArchivedFile(filename, messagesChan); err != nil {
			p.log.Error("error processing file", "filename", filename, "err", err)
		}
	}
	close(messagesChan)

	wgm.Wait()
}
