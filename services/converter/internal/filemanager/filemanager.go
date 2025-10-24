package filemanager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileManager struct {
	tempDir string
}

func New(tempDir string) (*FileManager, error) {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &FileManager{
		tempDir: tempDir,
	}, nil
}

func (fm *FileManager) CreateTempFile(jobID, extension string) (string, error) {
	// Create a unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d%s", jobID, timestamp, extension)
	filepath := filepath.Join(fm.tempDir, filename)

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	file.Close()

	return filepath, nil
}

func (fm *FileManager) SaveStreamToFile(reader io.Reader, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (fm *FileManager) OpenFile(filepath string) (*os.File, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (fm *FileManager) DeleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (fm *FileManager) CleanupTempFiles(maxAge time.Duration) error {
	entries, err := os.ReadDir(fm.tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := filepath.Join(fm.tempDir, entry.Name())
			if err := os.Remove(filepath); err != nil {
				// Log but don't fail on cleanup errors
				fmt.Printf("Failed to cleanup temp file %s: %v\n", filepath, err)
			}
		}
	}

	return nil
}

func (fm *FileManager) GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

func (fm *FileManager) GetFileSize(filepath string) (int64, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}