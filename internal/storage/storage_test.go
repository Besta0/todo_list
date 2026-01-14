package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
	apperrors "todolist/internal/errors"
	"todolist/internal/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestLoadCreatesNewFileWhenNotExists tests that Load returns an empty list when file doesn't exist
// Requirements: 5.2
func TestLoadCreatesNewFileWhenNotExists(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")

	// Create storage with non-existent file
	storage := NewFileStorage(nonExistentFile)

	// Load should return empty list without error
	taskList, err := storage.Load()
	if err != nil {
		t.Fatalf("Expected no error when loading non-existent file, got: %v", err)
	}

	// Verify empty list
	if taskList == nil {
		t.Fatal("Expected non-nil task list")
	}
	if len(taskList.Tasks) != 0 {
		t.Errorf("Expected empty task list, got %d tasks", len(taskList.Tasks))
	}
	if taskList.NextID != 1 {
		t.Errorf("Expected NextID to be 1, got %d", taskList.NextID)
	}
}

// TestLoadInvalidJSONReturnsError tests that Load returns error for invalid JSON
// Requirements: 5.4
func TestLoadInvalidJSONReturnsError(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	invalidJSONFile := filepath.Join(tempDir, "invalid.json")

	// Write invalid JSON to file
	invalidJSON := []byte(`{"tasks": [{"id": 1, "description": "test", "completed": false, "created_at": "invalid"}`)
	if err := os.WriteFile(invalidJSONFile, invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create storage
	storage := NewFileStorage(invalidJSONFile)

	// Load should return error
	_, err := storage.Load()
	if err == nil {
		t.Fatal("Expected error when loading invalid JSON, got nil")
	}

	// Verify it's an invalid JSON error
	if !errors.Is(err, apperrors.ErrInvalidJSON) {
		t.Errorf("Expected ErrInvalidJSON, got: %v", err)
	}
}

// TestLoadMalformedJSONReturnsError tests various malformed JSON scenarios
// Requirements: 5.4
func TestLoadMalformedJSONReturnsError(t *testing.T) {
	testCases := []struct {
		name        string
		jsonContent string
	}{
		{
			name:        "incomplete JSON",
			jsonContent: `{"tasks": [`,
		},
		{
			name:        "not JSON at all",
			jsonContent: `this is not json`,
		},
		{
			name:        "empty file",
			jsonContent: ``,
		},
		{
			name:        "wrong structure",
			jsonContent: `["array", "instead", "of", "object"]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.json")

			// Write malformed JSON
			if err := os.WriteFile(testFile, []byte(tc.jsonContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			storage := NewFileStorage(testFile)
			_, err := storage.Load()

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
			if !errors.Is(err, apperrors.ErrInvalidJSON) {
				t.Errorf("Expected ErrInvalidJSON for %s, got: %v", tc.name, err)
			}
		})
	}
}

// TestSaveFilePermissionError tests that Save returns error when file cannot be written
// Requirements: 5.5
func TestSaveFilePermissionError(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")

	// Create directory
	if err := os.Mkdir(readOnlyDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Make directory read-only
	if err := os.Chmod(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// Ensure cleanup restores permissions
	defer os.Chmod(readOnlyDir, 0755)

	// Try to save to read-only directory
	testFile := filepath.Join(readOnlyDir, "test.json")
	storage := NewFileStorage(testFile)

	taskList := &models.TaskList{
		Tasks: []models.Task{
			{
				ID:          1,
				Description: "Test task",
				Completed:   false,
				CreatedAt:   time.Now(),
			},
		},
		NextID: 2,
	}

	// Save should return error
	err := storage.Save(taskList)
	if err == nil {
		t.Fatal("Expected error when saving to read-only directory, got nil")
	}

	// Verify it's a storage write error
	if !errors.Is(err, apperrors.ErrStorageWrite) {
		t.Errorf("Expected ErrStorageWrite, got: %v", err)
	}
}

// TestSaveAndLoadRoundTrip tests that data can be saved and loaded correctly
// This is a basic integration test for the storage layer
func TestSaveAndLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	storage := NewFileStorage(testFile)

	// Create test data
	originalList := &models.TaskList{
		Tasks: []models.Task{
			{
				ID:          1,
				Description: "First task",
				Completed:   false,
				CreatedAt:   time.Now().Truncate(time.Second),
			},
			{
				ID:          2,
				Description: "Second task",
				Completed:   true,
				CreatedAt:   time.Now().Add(time.Hour).Truncate(time.Second),
			},
		},
		NextID: 3,
	}

	// Save
	if err := storage.Save(originalList); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load
	loadedList, err := storage.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify
	if loadedList.NextID != originalList.NextID {
		t.Errorf("NextID mismatch: expected %d, got %d", originalList.NextID, loadedList.NextID)
	}
	if len(loadedList.Tasks) != len(originalList.Tasks) {
		t.Fatalf("Task count mismatch: expected %d, got %d", len(originalList.Tasks), len(loadedList.Tasks))
	}

	for i, task := range loadedList.Tasks {
		orig := originalList.Tasks[i]
		if task.ID != orig.ID {
			t.Errorf("Task %d ID mismatch: expected %d, got %d", i, orig.ID, task.ID)
		}
		if task.Description != orig.Description {
			t.Errorf("Task %d Description mismatch: expected %s, got %s", i, orig.Description, task.Description)
		}
		if task.Completed != orig.Completed {
			t.Errorf("Task %d Completed mismatch: expected %v, got %v", i, orig.Completed, task.Completed)
		}
		if !task.CreatedAt.Equal(orig.CreatedAt) {
			t.Errorf("Task %d CreatedAt mismatch: expected %v, got %v", i, orig.CreatedAt, task.CreatedAt)
		}
	}
}

// Feature: todo-list-cli, Property 5: 持久化往返一致性
// Validates: Requirements 1.5, 3.3, 4.3, 5.1, 5.3
func TestProperty_PersistenceRoundTripConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for Task
	genTask := gopter.CombineGens(
		gen.Int(),
		gen.AnyString(),
		gen.Bool(),
		gen.TimeRange(time.Now().Add(-365*24*time.Hour), 24*365*time.Hour),
	).Map(func(values []interface{}) models.Task {
		return models.Task{
			ID:          values[0].(int),
			Description: values[1].(string),
			Completed:   values[2].(bool),
			CreatedAt:   values[3].(time.Time).Truncate(time.Second), // Truncate to second for JSON precision
		}
	})

	// Generator for TaskList
	genTaskList := gopter.CombineGens(
		gen.SliceOf(genTask),
		gen.IntRange(1, 1000),
	).Map(func(values []interface{}) *models.TaskList {
		tasks := values[0].([]models.Task)
		// Ensure Tasks is not nil
		if tasks == nil {
			tasks = []models.Task{}
		}
		return &models.TaskList{
			Tasks:  tasks,
			NextID: values[1].(int),
		}
	})

	properties.Property("保存然后加载应该产生等价的任务列表", prop.ForAll(
		func(originalList *models.TaskList) bool {
			// Create a temporary file for this test iteration
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.json")

			storage := NewFileStorage(testFile)

			// Save the original list
			if err := storage.Save(originalList); err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Load the list back
			loadedList, err := storage.Load()
			if err != nil {
				t.Logf("Load failed: %v", err)
				return false
			}

			// Verify NextID is the same
			if loadedList.NextID != originalList.NextID {
				t.Logf("NextID mismatch: expected %d, got %d", originalList.NextID, loadedList.NextID)
				return false
			}

			// Verify task count is the same
			if len(loadedList.Tasks) != len(originalList.Tasks) {
				t.Logf("Task count mismatch: expected %d, got %d", len(originalList.Tasks), len(loadedList.Tasks))
				return false
			}

			// Verify each task is equivalent
			for i, loadedTask := range loadedList.Tasks {
				origTask := originalList.Tasks[i]

				if loadedTask.ID != origTask.ID {
					t.Logf("Task %d ID mismatch: expected %d, got %d", i, origTask.ID, loadedTask.ID)
					return false
				}

				if loadedTask.Description != origTask.Description {
					t.Logf("Task %d Description mismatch: expected %q, got %q", i, origTask.Description, loadedTask.Description)
					return false
				}

				if loadedTask.Completed != origTask.Completed {
					t.Logf("Task %d Completed mismatch: expected %v, got %v", i, origTask.Completed, loadedTask.Completed)
					return false
				}

				// Compare timestamps (should be equal after JSON round-trip)
				if !loadedTask.CreatedAt.Equal(origTask.CreatedAt) {
					t.Logf("Task %d CreatedAt mismatch: expected %v, got %v", i, origTask.CreatedAt, loadedTask.CreatedAt)
					return false
				}
			}

			return true
		},
		genTaskList,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
