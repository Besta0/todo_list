package storage

import (
	"encoding/json"
	"errors"
	"os"
	apperrors "todolist/internal/errors"
	"todolist/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	Load() (*models.TaskList, error)
	Save(list *models.TaskList) error
}

// FileStorage implements Storage interface using file-based persistence
type FileStorage struct {
	filepath string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(filepath string) *FileStorage {
	return &FileStorage{
		filepath: filepath,
	}
}

// Load reads the task list from the file
func (fs *FileStorage) Load() (*models.TaskList, error) {
	// Read file content
	data, err := os.ReadFile(fs.filepath)
	if err != nil {
		// If file doesn't exist, return empty list
		if os.IsNotExist(err) {
			return &models.TaskList{
				Tasks:  []models.Task{},
				NextID: 1,
			}, nil
		}
		// Other read errors
		return nil, apperrors.WrapStorageReadError(errors.Join(apperrors.ErrStorageRead, err), fs.filepath)
	}

	// Parse JSON
	var taskList models.TaskList
	if err := json.Unmarshal(data, &taskList); err != nil {
		return nil, apperrors.WrapJSONError(errors.Join(apperrors.ErrInvalidJSON, err), fs.filepath)
	}

	// Ensure Tasks is not nil
	if taskList.Tasks == nil {
		taskList.Tasks = []models.Task{}
	}

	return &taskList, nil
}

// Save writes the task list to the file using atomic write
func (fs *FileStorage) Save(list *models.TaskList) error {
	// Serialize to JSON with indentation for readability
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return apperrors.WrapStorageWriteError(errors.Join(apperrors.ErrStorageWrite, err), fs.filepath)
	}

	// Use atomic write: write to temp file then rename
	tempFile := fs.filepath + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return apperrors.WrapStorageWriteError(errors.Join(apperrors.ErrStorageWrite, err), fs.filepath)
	}

	// Rename temp file to actual file (atomic operation)
	if err := os.Rename(tempFile, fs.filepath); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return apperrors.WrapStorageWriteError(errors.Join(apperrors.ErrStorageWrite, err), fs.filepath)
	}

	return nil
}
