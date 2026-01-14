package errors

import (
	"errors"
	"fmt"
)

// Core error types as defined in the design document

// Business logic errors
var (
	ErrEmptyDescription = errors.New("task description cannot be empty")
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidID        = errors.New("invalid task ID")
)

// Storage errors
var (
	ErrStorageRead  = errors.New("failed to read from storage")
	ErrStorageWrite = errors.New("failed to write to storage")
	ErrInvalidJSON  = errors.New("invalid JSON format")
)

// CLI errors
var (
	ErrInvalidCommand = errors.New("invalid command")
)

// Error wrapping utilities for adding context

// WrapWithContext wraps an error with additional context information
func WrapWithContext(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapStorageReadError wraps a storage read error with context
func WrapStorageReadError(err error, filepath string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to read from storage at %s: %w", filepath, err)
}

// WrapStorageWriteError wraps a storage write error with context
func WrapStorageWriteError(err error, filepath string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to write to storage at %s: %w", filepath, err)
}

// WrapJSONError wraps a JSON parsing error with context
func WrapJSONError(err error, filepath string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("invalid JSON format in %s: %w", filepath, err)
}

// WrapCommandError wraps a command execution error with context
func WrapCommandError(err error, command string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("command '%s' failed: %w", command, err)
}

// IsTaskNotFound checks if an error is ErrTaskNotFound
func IsTaskNotFound(err error) bool {
	return errors.Is(err, ErrTaskNotFound)
}

// IsInvalidID checks if an error is ErrInvalidID
func IsInvalidID(err error) bool {
	return errors.Is(err, ErrInvalidID)
}

// IsEmptyDescription checks if an error is ErrEmptyDescription
func IsEmptyDescription(err error) bool {
	return errors.Is(err, ErrEmptyDescription)
}

// IsStorageError checks if an error is a storage-related error
func IsStorageError(err error) bool {
	return errors.Is(err, ErrStorageRead) || errors.Is(err, ErrStorageWrite)
}

// IsInvalidJSON checks if an error is ErrInvalidJSON
func IsInvalidJSON(err error) bool {
	return errors.Is(err, ErrInvalidJSON)
}

// IsInvalidCommand checks if an error is ErrInvalidCommand
func IsInvalidCommand(err error) bool {
	return errors.Is(err, ErrInvalidCommand)
}
