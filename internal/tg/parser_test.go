package tg

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meesooqa/tgtag/pkg/models"
)

func TestTgArchivedHTMLParser_ParseFile(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	parser := NewTgArchivedHTMLParser(logger, "")
	testFile := "testdata/test.html"

	messagesChan := make(chan models.Message, 10)
	err := parser.ParseFile(testFile, messagesChan)
	require.NoError(t, err)

	require.Equal(t, 3, len(messagesChan), "Ожидается, что будет 3 сообщения, полученных из HTML")

	var messages []models.Message
	for i := 0; i < 3; i++ {
		msg := <-messagesChan
		messages = append(messages, msg)
	}

	var msg2203, msg2204, msg3217 *models.Message
	for i := range messages {
		switch messages[i].MessageID {
		case "message2203":
			msg2203 = &messages[i]
		case "message2204":
			msg2204 = &messages[i]
		case "message3217":
			msg3217 = &messages[i]
		}
	}

	require.NotNil(t, msg2203, "Сообщение с ID 'message2203' должно быть найдено")
	require.NotNil(t, msg2204, "Сообщение с ID 'message2204' должно быть найдено")
	require.NotNil(t, msg3217, "Сообщение с ID 'message3217' должно быть найдено")

	fixedZone := time.FixedZone("UTC+03:00", 3*60*60)

	expectedTime2203 := time.Date(2024, time.November, 21, 19, 20, 37, 0, fixedZone)
	assert.Equal(t, expectedTime2203, msg2203.Datetime, "Некорректная дата для message2203")
	assert.ElementsMatch(t, []string{"shy", "booba"}, msg2203.Tags, "Некорректные теги для message2203")

	expectedTime2204 := time.Date(2024, time.November, 21, 19, 20, 37, 0, fixedZone)
	assert.Equal(t, expectedTime2204, msg2204.Datetime, "Некорректная дата для message2204")
	assert.ElementsMatch(t, []string{"shy", "stare", "todo", "ginger"}, msg2204.Tags, "Некорректные теги для message2204")

	expectedTime3217 := time.Date(2025, time.January, 29, 11, 52, 44, 0, fixedZone)
	assert.Equal(t, expectedTime3217, msg3217.Datetime, "Некорректная дата для message3217")
	assert.ElementsMatch(t, []string{"where", "booba", "slontar4"}, msg3217.Tags, "Некорректные теги для message3217")
}

func TestParseTZOffset(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"UTC+03:00", 3 * 3600, false},
		{"UTC-05:30", -5*3600 - 30*60, false},
		{"UTC+00:00", 0, false},
		{"UTC-00:45", -45 * 60, false},
		{"UTC+12:15", 12*3600 + 15*60, false},
		{"UTC-11:59", -11*3600 - 59*60, false},
		{"UTC+99:99", 0, true}, // Нереальная зона
		{"GMT+03:00", 0, true}, // Неподдерживаемый формат
		{"UTC+3", 0, true},     // Неполный формат
		{"UTC+03:XX", 0, true}, // Ошибка в минутах
		{"random", 0, true},    // Полностью некорректный ввод
		{"UTC+15:00", 0, true}, // Превышает максимум (14:00)
		{"UTC-14:01", 0, true}, // Превышает минимум (-14:00)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseTZOffset(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("parseTZOffset(%q) error = %v, expected error = %v", tt.input, err, tt.hasError)
			}
			if result != tt.expected {
				t.Errorf("parseTZOffset(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractFirstSubfolder(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	parser := NewTgArchivedHTMLParser(logger, "")

	testCases := []struct {
		path     string
		base     string
		expected string
	}{
		{"var/data/sub/file.txt", "var/data", "sub"},
		{"/home/gpt/var/data/sub1/sub2/file1.txt", "var/data", "sub1"},
		{"foo/bar/sub/file.txt", "foo/bar", "sub"},
		{"var/data/", "var/data", ""},
		{"some/var/data", "var/data", ""},
		{"prefix/var/data/subfolder", "var/data", "subfolder"},
		{"var/data/subfolder/file.txt", "var/data/", "subfolder"},
	}

	for _, tc := range testCases {
		result := parser.extractFirstSubfolder(tc.path, tc.base)
		assert.Equal(t, tc.expected, result)
	}
}
