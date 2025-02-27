package proc

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/meesooqa/tgtag/internal/proc/mocks"
)

func TestProcessor_ProcessFile_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	fService := &mocks.ServiceMock{}
	fRepo := &mocks.RepositoryMock{}
	processor := NewProcessor(logger, fService, fRepo)

	filesChan := make(chan string, 1)
	filesChan <- "file1.txt"
	close(filesChan)

	var wg sync.WaitGroup
	wg.Add(1)
	processor.ProcessFile(filesChan, &wg)
	wg.Wait()

	assert.Equal(t, 1, fService.CallCount, "Ожидается, что ParseArchivedFile будет вызван один раз")
	assert.Equal(t, 1, len(fRepo.UpsertCalls), "Ожидается, что одно сообщение будет обработано")
	assert.Equal(t, "file1.txt", fRepo.UpsertCalls[0].MessageID)
}

func TestProcessor_ProcessFile_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	parseErr := errors.New("db upsert error")
	fService := &mocks.ServiceMock{
		Err: parseErr,
	}
	fRepo := &mocks.RepositoryMock{}
	processor := NewProcessor(logger, fService, fRepo)

	filesChan := make(chan string, 1)
	filesChan <- "file2.txt"
	close(filesChan)

	var wg sync.WaitGroup
	wg.Add(1)
	processor.ProcessFile(filesChan, &wg)
	wg.Wait()

	assert.Equal(t, 1, fService.CallCount, "Ожидается, что ParseArchivedFile будет вызван один раз")
	assert.Equal(t, 1, len(fRepo.UpsertCalls), "Ожидается, что одно сообщение будет обработано")

	// Получаем JSON-лог
	logOutput := buf.String()
	// Десериализуем JSON
	var logMap map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &logMap)
	assert.NoError(t, err, "Должен быть корректный JSON")

	assert.Equal(t, "error processing file", logMap["msg"], "Ожидается, что ошибка обработки сообщения будет залогирована")
	assert.Equal(t, "file2.txt", logMap["filename"])
}
