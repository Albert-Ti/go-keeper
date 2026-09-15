package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Credentials struct {
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type FileStorage struct {
	mu    sync.Mutex
	path  string
	creds Credentials
}

func NewFileStorage() (*FileStorage, error) {
	// macOS:
	// ~/Library/Application Support/go_keeper/local_storage.json
	//
	// Linux:
	// ~/.config/go_keeper/local_storage.json
	//
	// Windows:
	// %AppData%\go_keeper\local_storage.json

	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(dir, "go_keeper")

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "local_storage.json")

	fs := &FileStorage{
		path: path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Файл ещё не существует.
			return fs, nil
		}

		return nil, err
	}

	// Файл существует, но пустой.
	if len(data) == 0 {
		return fs, nil
	}

	if err := json.Unmarshal(data, &fs.creds); err != nil {
		return nil, err
	}

	return fs, nil
}

func (f *FileStorage) Credentials() Credentials {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.creds
}

func (f *FileStorage) SaveCredentials(creds Credentials) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.creds = creds

	return f.flush()
}

func (f *FileStorage) Clear() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.creds = Credentials{}

	return f.flush()
}

// flush перезаписывает файл целиком.
func (f *FileStorage) flush() error {
	data, err := json.MarshalIndent(f.creds, "", "  ")
	if err != nil {
		return err
	}

	// Атомарная запись:
	// сначала пишем во временный файл,
	// потом заменяем основной.
	tmp := f.path + ".tmp"

	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, f.path)
}
