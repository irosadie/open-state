package infrastructure

import (
	"os"
	"path/filepath"

	"github.com/vibecoding-starter/api/internal/domain/services"
)

type LocalStorageService struct {
	baseURL string
}

func NewLocalStorageService() services.StorageService {
	baseURL := os.Getenv("STORAGE_BASE_URL")
	if baseURL == "" {
		baseURL = "/storage"
	}
	return &LocalStorageService{baseURL: baseURL}
}

func (s *LocalStorageService) GetURL(path string) string {
	if path == "" {
		return ""
	}
	return s.baseURL + "/" + filepath.Clean(path)
}
