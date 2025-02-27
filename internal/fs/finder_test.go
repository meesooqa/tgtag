package fs

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFiles_File(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	finder := NewFinder(logger)

	tmpFile, err := os.CreateTemp("", "testfile_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	filesChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go finder.FindFiles(tmpFile.Name(), filesChan, &wg)

	var files []string
	for file := range filesChan {
		files = append(files, file)
	}
	wg.Wait()

	assert.Equal(t, []string{tmpFile.Name()}, files)
}

func TestFindFiles_Directory(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	finder := NewFinder(logger)

	tempDir, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	err = os.WriteFile(file1Path, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2Path, []byte("content2"), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)
	subFilePath := filepath.Join(subDir, "subfile.txt")
	err = os.WriteFile(subFilePath, []byte("subcontent"), 0644)
	require.NoError(t, err)

	filesChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go finder.FindFiles(tempDir, filesChan, &wg)

	var files []string
	for file := range filesChan {
		files = append(files, file)
	}
	wg.Wait()

	expectedFiles := []string{file1Path, file2Path, subFilePath}
	assert.ElementsMatch(t, expectedFiles, files)
}

func TestFindFiles_NonExistent(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	finder := NewFinder(logger)

	nonExistentPath := "/nonexistentpath_123456789"
	filesChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go finder.FindFiles(nonExistentPath, filesChan, &wg)

	var files []string
	for file := range filesChan {
		files = append(files, file)
	}
	wg.Wait()

	assert.Empty(t, files)
}

func TestFindFiles_NonExistent_Logs(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	finder := NewFinder(logger)

	nonExistentPath := "/nonexistentpath_123456789"
	filesChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go finder.FindFiles(nonExistentPath, filesChan, &wg)

	// Считываем канал (он должен быть закрыт)
	for range filesChan {
	}
	wg.Wait()

	assert.Contains(t, buf.String(), "Can't stat file", "Ожидаемое сообщение об ошибке 'Can't stat file' не найдено")
}

func TestFindFiles_DirectoryWithInaccessibleSubDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Пропускаем тест для недоступной поддиректории на Windows")
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	finder := NewFinder(logger)

	tempDir, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	accessibleFile := filepath.Join(tempDir, "file1.txt")
	err = os.WriteFile(accessibleFile, []byte("content"), 0644)
	require.NoError(t, err)

	inaccessibleDir := filepath.Join(tempDir, "inaccessible")
	err = os.Mkdir(inaccessibleDir, 0755)
	require.NoError(t, err)
	// Убираем все права доступа, чтобы симулировать ошибку чтения
	err = os.Chmod(inaccessibleDir, 0000)
	require.NoError(t, err)

	filesChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go finder.FindFiles(tempDir, filesChan, &wg)

	var files []string
	for file := range filesChan {
		files = append(files, file)
	}
	wg.Wait()

	assert.Contains(t, files, accessibleFile)
	assert.Contains(t, buf.String(), "error while walking", "Ожидаемое сообщение об ошибке при обходе недоступной директории не найдено")

	err = os.Chmod(inaccessibleDir, 0755)
	require.NoError(t, err)
}
