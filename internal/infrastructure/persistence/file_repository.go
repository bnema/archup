package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// FileRepository implements the InstallationRepository port using filesystem storage
type FileRepository struct {
	basePath string
}

// NewFileRepository creates a new file-based repository
// basePath is the directory where installation state files will be stored
func NewFileRepository(basePath string) (*FileRepository, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository base path: %w", err)
	}

	return &FileRepository{
		basePath: basePath,
	}, nil
}

// Save persists the installation state to a file
func (fr *FileRepository) Save(ctx context.Context, installationID string, state string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := filepath.Join(fr.basePath, installationID+".state")

	// Write state to file
	if err := os.WriteFile(filePath, []byte(state), 0644); err != nil {
		return fmt.Errorf("failed to save installation state: %w", err)
	}

	return nil
}

// Load retrieves the installation state from a file
func (fr *FileRepository) Load(ctx context.Context, installationID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	filePath := filepath.Join(fr.basePath, installationID+".state")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("installation state not found: %w", err)
		}
		return "", fmt.Errorf("failed to load installation state: %w", err)
	}

	return string(data), nil
}

// Exists checks if an installation state exists
func (fr *FileRepository) Exists(ctx context.Context, installationID string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	filePath := filepath.Join(fr.basePath, installationID+".state")

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check installation existence: %w", err)
}
