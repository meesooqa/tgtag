package mocks

import "github.com/meesooqa/tgtag/pkg/models"

type ServiceMock struct {
	CallCount int
	Err       error
}

func (fs *ServiceMock) ParseArchivedFile(filename string, messagesChan chan<- models.Message) error {
	fs.CallCount++
	messagesChan <- models.Message{
		MessageID: filename,
	}
	return fs.Err
}
