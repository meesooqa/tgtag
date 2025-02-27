package tg

import (
	"log/slog"

	"github.com/meesooqa/tgtag/internal/config"
	"github.com/meesooqa/tgtag/pkg/models"
)

type Service interface {
	ParseArchivedFile(filename string, messagesChan chan<- models.Message) error
}

type TgService struct {
	log    *slog.Logger
	parser Parser
}

func NewService(log *slog.Logger, conf *config.SystemConfig) *TgService {
	return &TgService{
		log:    log,
		parser: NewTgArchivedHTMLParser(log, conf.DataPath),
	}
}

func (s *TgService) ParseArchivedFile(filename string, messagesChan chan<- models.Message) error {
	return s.parser.ParseFile(filename, messagesChan)
}
