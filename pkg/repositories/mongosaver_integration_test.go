package repositories

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoURI string

// TestMain поднимает MongoDB в контейнере перед тестами и удаляет после.
func TestMain(m *testing.M) {
	// Запускаем MongoDB в Docker через testcontainers-go
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(10 * time.Second),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start MongoDB container: %v", err)
	}

	// Получаем хост и порт
	mappedPort, err := mongoC.MappedPort(ctx, "27017")
	if err != nil {
		log.Fatalf("Failed to get mapped port: %v", err)
	}

	hostIP, err := mongoC.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	// Формируем URI для подключения
	mongoURI = fmt.Sprintf("mongodb://%s:%s", hostIP, mappedPort.Port())

	// Запускаем тесты
	code := m.Run()

	// Останавливаем контейнер
	if err := mongoC.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate MongoDB container: %v", err)
	}

	os.Exit(code)
}

// TestSaver_Integration проверяет реальную вставку и обновление в MongoDB.
func TestSaver_Integration(t *testing.T) {
	ctx := context.Background()

	// Подключаемся к MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Очищаем тестовую коллекцию
	db := client.Database("testdb")
	collection := db.Collection("messages")
	if err := collection.Drop(ctx); err != nil {
		t.Fatalf("Failed to drop collection: %v", err)
	}

	// Создаём Saver
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	svr := newSaver(logger, collection, 2, 100*time.Millisecond, 10)

	now := time.Now()
	doc1 := bson.M{
		"uuid":     "msg1",
		"tags":     []string{"tag1", "tag2"},
		"datetime": now,
	}
	doc2 := bson.M{
		"uuid":     "msg2",
		"tags":     []string{"tag3"},
		"datetime": now.Add(time.Minute),
	}

	// Добавляем документы
	if err := svr.Save(doc1); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if err := svr.Save(doc2); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Ждём, пока данные запишутся
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что документы вставлены
	var result1, result2 bson.M
	if err := collection.FindOne(ctx, bson.M{"uuid": "msg1"}).Decode(&result1); err != nil {
		t.Fatalf("Failed to find msg1: %v", err)
	}
	if err := collection.FindOne(ctx, bson.M{"uuid": "msg2"}).Decode(&result2); err != nil {
		t.Fatalf("Failed to find msg2: %v", err)
	}

	// Проверяем обновление: добавляем теги
	doc1Update := bson.M{
		"uuid":     "msg1",
		"tags":     []string{"tag1", "tag2", "new_tag"},
		"datetime": now.Add(time.Hour),
	}

	if err := svr.Save(doc1Update); err != nil {
		t.Fatalf("Update Save failed: %v", err)
	}

	// Ждём обновления
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что msg1 обновился
	var updatedResult1 bson.M
	if err := collection.FindOne(ctx, bson.M{"uuid": "msg1"}).Decode(&updatedResult1); err != nil {
		t.Fatalf("Failed to find updated msg1: %v", err)
	}

	// Проверяем, что tags обновился
	expectedTags := []string{"tag1", "tag2", "new_tag"}
	assert.ElementsMatch(t, updatedResult1["tags"], expectedTags)

	// Завершаем работу Saver
	svr.Close()
}
