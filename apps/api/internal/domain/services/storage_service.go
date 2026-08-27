package services

// StorageService defines the contract for file/object storage operations.
type StorageService interface {
	GetURL(path string) string
}
