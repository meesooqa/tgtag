package tg

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"

	"github.com/meesooqa/tgtag/pkg/models"
)

type Parser interface {
	ParseFile(filename string, messagesChan chan<- models.Message) error
}

type TgArchivedHTMLParser struct {
	log     *slog.Logger
	baseDir string
}

func NewTgArchivedHTMLParser(log *slog.Logger, baseDir string) *TgArchivedHTMLParser {
	return &TgArchivedHTMLParser{
		log:     log,
		baseDir: baseDir,
	}
}

func (p *TgArchivedHTMLParser) ParseFile(filename string, messagesChan chan<- models.Message) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return err
	}

	group := p.obtainGroup(filename)

	doc.Find("div.message.default").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			return
		}

		dateStr, exists := s.Find("div.pull_right.date.details").Attr("title")
		if !exists {
			return
		}
		// "21.11.2024 19:20:37 UTC+03:00"
		parts := strings.Split(dateStr, " ")
		if len(parts) < 3 {
			return
		}

		dateTimeStr := parts[0] + " " + parts[1] // "21.11.2024 19:20:37"
		parsedTime, err := time.Parse("02.01.2006 15:04:05", dateTimeStr)
		if err != nil {
			return
		}

		tzStr := parts[2] // "UTC+03:00"
		offset, err := parseTZOffset(tzStr)
		if err != nil {
			return
		}
		loc := time.FixedZone(tzStr, offset)

		datetime := time.Date(
			parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
			parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
			0, loc,
		)

		var tags []string
		s.Find("div.text a").Each(func(i int, a *goquery.Selection) {
			text := strings.TrimPrefix(a.Text(), "#")
			tags = append(tags, text)
		})

		messagesChan <- models.Message{
			UUID:      p.obtainUUID(id, group),
			MessageID: id,
			Datetime:  datetime,
			Group:     group,
			Tags:      tags,
		}
	})

	return nil
}

func (p *TgArchivedHTMLParser) obtainUUID(messageId, group string) string {
	input := messageId + group

	namespace := uuid.NameSpaceURL
	return uuid.NewSHA1(namespace, []byte(input)).String()
}

func (p *TgArchivedHTMLParser) obtainGroup(path string) string {
	return p.extractFirstSubfolder(path, p.baseDir)
}

// parseTZOffset парсит строку часового пояса формата "UTC±HH:MM" и возвращает смещение в секундах.
func parseTZOffset(offsetStr string) (int, error) {
	re := regexp.MustCompile(`^UTC([+-])(\d{2}):(\d{2})$`)
	matches := re.FindStringSubmatch(offsetStr)
	if matches == nil {
		return 0, fmt.Errorf("invalid timezone format: %s", offsetStr)
	}

	sign := matches[1]
	hours, _ := strconv.Atoi(matches[2])
	minutes, _ := strconv.Atoi(matches[3])

	// Проверяем диапазон значений
	if hours > 14 || (hours == 14 && minutes > 0) || minutes >= 60 {
		return 0, fmt.Errorf("invalid timezone values: %s", offsetStr)
	}

	totalSeconds := hours*3600 + minutes*60
	if sign == "-" {
		totalSeconds = -totalSeconds
	}

	return totalSeconds, nil
}

// extractFirstSubfolder возвращает первую подпапку, следующую за каталогом base
func (p *TgArchivedHTMLParser) extractFirstSubfolder(path string, base string) string {
	// Нормализуем базовый путь, удаляя возможные ведущие и завершающие слэши.
	base = strings.Trim(base, "/")
	baseParts := strings.Split(base, "/")
	pathParts := strings.Split(path, "/")

	// Ищем последовательное вхождение baseParts в pathParts.
	// Вычитаем единицу, чтобы после последовательности осталась хотя бы одна часть.
	for i := 0; i <= len(pathParts)-len(baseParts)-1; i++ {
		match := true
		for j, bp := range baseParts {
			if pathParts[i+j] != bp {
				match = false
				break
			}
		}
		if match {
			// Проверяем, что следующая часть существует и не пуста.
			if i+len(baseParts) < len(pathParts) && pathParts[i+len(baseParts)] != "" {
				return pathParts[i+len(baseParts)]
			}
		}
	}
	return ""
}
