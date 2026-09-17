package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type storageKey string

const (
	emailKey        storageKey = "email"
	accessTokenKey  storageKey = "access_token"
	refreshTokenKey storageKey = "refresh_token"
)

type fileRecord struct {
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type FileStorage struct {
	mu     sync.Mutex
	path   string
	record fileRecord
}

func NewFileStorage() (*FileStorage, error) {
	// ~/Library/Application Support/go_keeper/local_storage.json - macOS
	// ~/.local/share/go_keeper/local_storage.json - Linux
	// %AppData%\go_keeper\local_storage.json - Win
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "go_keeper")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "local_storage.json")

	fs := &FileStorage{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// файла ещё нет
			return fs, nil
		}
		return nil, err
	}

	// пустой файл
	if len(data) == 0 {
		return fs, nil
	}

	if err := json.Unmarshal(data, &fs.record); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStorage) Get(key storageKey) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch key {
	case "email":
		return f.record.Email
	case "access_token":
		return f.record.AccessToken
	case "refresh_token":
		return f.record.RefreshToken
	}
	return ""
}

func (f *FileStorage) Set(key storageKey, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch key {
	case "email":
		f.record.Email = value
	case "access_token":
		f.record.AccessToken = value
	case "refresh_token":
		f.record.RefreshToken = value
	default:
		return errors.New("unknown key: " + string(key))
	}
	return f.flush()
}

// flush перезаписывает файл целиком.
func (f *FileStorage) flush() error {
	data, err := json.MarshalIndent(f.record, "", "  ")
	if err != nil {
		return err
	}
	// атомарная запись: пишем во временный файл, потом переименовываем
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}
