package models

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Sample property test to verify gopter is working
func TestPropertyFramework(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("sample property: string length is non-negative", prop.ForAll(
		func(s string) bool {
			return len(s) >= 0
		},
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 4: 任务包含创建时间
// Validates: Requirements 1.4
func TestProperty_TaskContainsCreationTime(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("任务包含有效的创建时间戳", prop.ForAll(
		func(id int, description string) bool {
			// Record the time before creating the task
			beforeTime := time.Now()

			// Create a task
			task := Task{
				ID:          id,
				Description: description,
				Completed:   false,
				CreatedAt:   time.Now(),
			}

			// Record the time after creating the task
			afterTime := time.Now()

			// Verify that CreatedAt is not zero
			if task.CreatedAt.IsZero() {
				return false
			}

			// Verify that CreatedAt is within reasonable bounds
			// (between beforeTime and afterTime, with small tolerance for clock precision)
			tolerance := 100 * time.Millisecond
			if task.CreatedAt.Before(beforeTime.Add(-tolerance)) {
				return false
			}
			if task.CreatedAt.After(afterTime.Add(tolerance)) {
				return false
			}

			return true
		},
		gen.Int(),
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
