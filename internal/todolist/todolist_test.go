package todolist

import (
	"strings"
	"testing"
	apperrors "todolist/internal/errors"
	"todolist/internal/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Test file for core business logic
func TestPlaceholder(t *testing.T) {
	// Placeholder test to verify test framework is working
	t.Log("Test framework is set up correctly")
}

// Basic integration test to verify core functionality
func TestTodoListBasicOperations(t *testing.T) {
	// Create storage and todolist
	storage := &mockStorage{data: nil}
	tl, err := NewTodoList(storage)
	if err != nil {
		t.Fatalf("Failed to create TodoList: %v", err)
	}

	// Test AddTask
	task1, err := tl.AddTask("Test task 1")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if task1.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", task1.ID)
	}
	if task1.Description != "Test task 1" {
		t.Errorf("Expected description 'Test task 1', got '%s'", task1.Description)
	}
	if task1.Completed {
		t.Error("New task should not be completed")
	}

	// Test AddTask with empty description
	_, err = tl.AddTask("   ")
	if err != apperrors.ErrEmptyDescription {
		t.Errorf("Expected apperrors.ErrEmptyDescription, got %v", err)
	}

	// Add another task
	task2, err := tl.AddTask("Test task 2")
	if err != nil {
		t.Fatalf("Failed to add second task: %v", err)
	}
	if task2.ID != 2 {
		t.Errorf("Expected task ID 2, got %d", task2.ID)
	}

	// Test ListTasks
	tasks := tl.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Test CompleteTask
	err = tl.CompleteTask(1)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Verify task is completed
	tasks = tl.ListTasks()
	if !tasks[0].Completed {
		t.Error("Task 1 should be completed")
	}

	// Test CompleteTask with invalid ID
	err = tl.CompleteTask(999)
	if err != apperrors.ErrTaskNotFound {
		t.Errorf("Expected apperrors.ErrTaskNotFound, got %v", err)
	}

	// Test DeleteTask
	err = tl.DeleteTask(1)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task is deleted
	tasks = tl.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task after deletion, got %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("Expected remaining task to have ID 2, got %d", tasks[0].ID)
	}

	// Test DeleteTask with invalid ID
	err = tl.DeleteTask(999)
	if err != apperrors.ErrTaskNotFound {
		t.Errorf("Expected apperrors.ErrTaskNotFound, got %v", err)
	}
}

// mockStorage is a simple in-memory storage for testing
type mockStorage struct {
	data *models.TaskList
}

func (ms *mockStorage) Load() (*models.TaskList, error) {
	if ms.data == nil {
		return &models.TaskList{
			Tasks:  []models.Task{},
			NextID: 1,
		}, nil
	}
	return ms.data, nil
}

func (ms *mockStorage) Save(list *models.TaskList) error {
	// Deep copy to simulate persistence
	tasks := make([]models.Task, len(list.Tasks))
	copy(tasks, list.Tasks)
	ms.data = &models.TaskList{
		Tasks:  tasks,
		NextID: list.NextID,
	}
	return nil
}

// Feature: todo-list-cli, Property 1: 添加任务增加列表长度
// For any valid (non-empty) task description, adding it to the task list should increase the list length by 1
// Validates: Requirements 1.1
func TestProperty_AddTaskIncreasesLength(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("adding valid task increases list length by 1",
		prop.ForAll(
			func(description string) bool {
				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Get initial length
				initialLength := len(tl.ListTasks())

				// Add task
				task, err := tl.AddTask(description)
				if err != nil {
					return false
				}

				// Verify task was created
				if task == nil {
					return false
				}

				// Get new length
				newLength := len(tl.ListTasks())

				// Verify length increased by exactly 1
				return newLength == initialLength+1
			},
			// Generate non-empty strings (after trimming whitespace)
			gen.AnyString().SuchThat(func(s string) bool {
				return strings.TrimSpace(s) != ""
			}),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 2: 空白任务被拒绝
// For any string composed entirely of whitespace characters, attempting to add it as a task should be rejected and the task list should remain unchanged
// Validates: Requirements 1.2
func TestProperty_BlankTasksRejected(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("blank tasks are rejected and list remains unchanged",
		prop.ForAll(
			func(whitespaceStr string) bool {
				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Get initial length
				initialLength := len(tl.ListTasks())

				// Attempt to add blank task
				task, err := tl.AddTask(whitespaceStr)

				// Verify error is returned
				if err != apperrors.ErrEmptyDescription {
					return false
				}

				// Verify no task was created
				if task != nil {
					return false
				}

				// Verify list length unchanged
				newLength := len(tl.ListTasks())
				if newLength != initialLength {
					return false
				}

				return true
			},
			// Generate strings composed entirely of whitespace characters
			gen.OneGenOf(
				gen.Const(""),         // Empty string
				gen.Const(" "),        // Single space
				gen.Const("  "),       // Multiple spaces
				gen.Const("\t"),       // Tab
				gen.Const("\n"),       // Newline
				gen.Const("\r"),       // Carriage return
				gen.Const("   \t  "),  // Mixed spaces and tabs
				gen.Const("\n\n"),     // Multiple newlines
				gen.Const(" \t\n\r "), // All whitespace types
				gen.Const("     "),    // Many spaces
				gen.Const("\t\t\t"),   // Many tabs
			),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 3: 任务 ID 唯一性
// For any sequence of tasks, each task added to the list should have a unique ID, and IDs should be incrementing
// Validates: Requirements 1.3
func TestProperty_TaskIDUniqueness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("each task has unique and incrementing ID",
		prop.ForAll(
			func(count int) bool {
				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Track all IDs we've seen
				seenIDs := make(map[int]bool)
				var previousID int = 0

				// Add tasks and verify ID uniqueness and incrementing
				for i := 0; i < count; i++ {
					desc := "Task " + string(rune('A'+i%26))
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}

					// Check ID is unique
					if seenIDs[task.ID] {
						return false // Duplicate ID found
					}
					seenIDs[task.ID] = true

					// Check ID is incrementing (greater than previous)
					if task.ID <= previousID {
						return false // ID not incrementing
					}
					previousID = task.ID
				}

				// Verify all tasks in the list have unique IDs
				listedTasks := tl.ListTasks()
				listIDs := make(map[int]bool)
				for _, task := range listedTasks {
					if listIDs[task.ID] {
						return false // Duplicate ID in list
					}
					listIDs[task.ID] = true
				}

				return true
			},
			// Generate a count between 1 and 20
			gen.IntRange(1, 20),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 6: 列表操作返回所有任务
// For any task list state, calling ListTasks should return all tasks in the list with complete information
// Validates: Requirements 2.1, 2.2
func TestProperty_ListTasksReturnsAllTasks(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("ListTasks returns all tasks with complete information",
		prop.ForAll(
			func(descriptions []string) bool {
				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks and track them
				addedTasks := make([]*models.Task, 0, len(descriptions))
				for _, desc := range descriptions {
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
					addedTasks = append(addedTasks, task)
				}

				// Get list of tasks
				listedTasks := tl.ListTasks()

				// Verify count matches
				if len(listedTasks) != len(addedTasks) {
					return false
				}

				// Verify each added task appears in the list with complete information
				for i, addedTask := range addedTasks {
					listedTask := listedTasks[i]

					// Check all fields are preserved
					if listedTask.ID != addedTask.ID {
						return false
					}
					if listedTask.Description != addedTask.Description {
						return false
					}
					if listedTask.Completed != addedTask.Completed {
						return false
					}
					// Check CreatedAt is preserved (allowing for minor time differences due to copying)
					if !listedTask.CreatedAt.Equal(addedTask.CreatedAt) {
						return false
					}
				}

				return true
			},
			// Generate slices of non-empty strings
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for empty list edge case
// Tests that an empty list returns an empty slice (not nil)
// Validates: Requirements 2.3
func TestEmptyListReturnsEmptySlice(t *testing.T) {
	// Create storage and todolist with no tasks
	storage := &mockStorage{data: nil}
	tl, err := NewTodoList(storage)
	if err != nil {
		t.Fatalf("Failed to create TodoList: %v", err)
	}

	// Get list of tasks from empty list
	tasks := tl.ListTasks()

	// Verify it returns an empty slice, not nil
	if tasks == nil {
		t.Error("ListTasks should return an empty slice, not nil")
	}

	// Verify the length is 0
	if len(tasks) != 0 {
		t.Errorf("Expected empty list to have length 0, got %d", len(tasks))
	}

	// Verify we can safely iterate over it
	count := 0
	for range tasks {
		count++
	}
	if count != 0 {
		t.Errorf("Expected 0 iterations over empty list, got %d", count)
	}
}

// Feature: todo-list-cli, Property 7: 任务按创建时间排序
// For any task list, ListTasks should return tasks sorted by creation time in ascending order
// Validates: Requirements 2.4
func TestProperty_TasksSortedByCreationTime(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("tasks are sorted by creation time in ascending order",
		prop.ForAll(
			func(descriptions []string) bool {
				// Skip empty lists
				if len(descriptions) == 0 {
					return true
				}

				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks
				for _, desc := range descriptions {
					_, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
				}

				// Get list of tasks
				listedTasks := tl.ListTasks()

				// Verify tasks are sorted by creation time (ascending)
				for i := 0; i < len(listedTasks)-1; i++ {
					currentTask := listedTasks[i]
					nextTask := listedTasks[i+1]

					// Current task's creation time should be before or equal to next task's creation time
					if currentTask.CreatedAt.After(nextTask.CreatedAt) {
						return false // Not sorted correctly
					}
				}

				return true
			},
			// Generate slices of non-empty strings
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			).SuchThat(func(s []string) bool {
				// Generate lists with at least 2 tasks to make sorting meaningful
				return len(s) >= 2
			}),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 8: 完成任务更新状态
// For any existing task ID, calling CompleteTask should set the task's Completed field to true
// Validates: Requirements 3.1
func TestProperty_CompleteTaskUpdatesStatus(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("completing a task sets its Completed field to true",
		prop.ForAll(
			func(descriptions []string, taskIndex int) bool {
				// Skip empty lists
				if len(descriptions) == 0 {
					return true
				}

				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks
				addedTasks := make([]*models.Task, 0, len(descriptions))
				for _, desc := range descriptions {
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
					addedTasks = append(addedTasks, task)
				}

				// Select a task to complete (using modulo to ensure valid index)
				selectedTask := addedTasks[taskIndex%len(addedTasks)]

				// Verify task is initially not completed
				if selectedTask.Completed {
					return false
				}

				// Complete the task
				err = tl.CompleteTask(selectedTask.ID)
				if err != nil {
					return false
				}

				// Get the updated task list
				listedTasks := tl.ListTasks()

				// Find the completed task in the list
				var completedTask *models.Task
				for i := range listedTasks {
					if listedTasks[i].ID == selectedTask.ID {
						completedTask = &listedTasks[i]
						break
					}
				}

				// Verify task was found
				if completedTask == nil {
					return false
				}

				// Verify the task is now marked as completed
				if !completedTask.Completed {
					return false
				}

				// Verify other fields remain unchanged
				if completedTask.ID != selectedTask.ID {
					return false
				}
				if completedTask.Description != selectedTask.Description {
					return false
				}
				if !completedTask.CreatedAt.Equal(selectedTask.CreatedAt) {
					return false
				}

				return true
			},
			// Generate slices of non-empty strings (at least 1 task)
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			).SuchThat(func(s []string) bool {
				return len(s) >= 1
			}),
			// Generate a task index to select which task to complete
			gen.IntRange(0, 1000),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 9: 完成操作幂等性
// For any task, calling CompleteTask multiple times should produce the same result (task remains completed)
// Validates: Requirements 3.4
func TestProperty_CompleteTaskIdempotent(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("completing a task multiple times is idempotent",
		prop.ForAll(
			func(descriptions []string, taskIndex int, repeatCount int) bool {
				// Skip empty lists
				if len(descriptions) == 0 {
					return true
				}

				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks
				addedTasks := make([]*models.Task, 0, len(descriptions))
				for _, desc := range descriptions {
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
					addedTasks = append(addedTasks, task)
				}

				// Select a task to complete (using modulo to ensure valid index)
				selectedTask := addedTasks[taskIndex%len(addedTasks)]

				// Complete the task the first time
				err = tl.CompleteTask(selectedTask.ID)
				if err != nil {
					return false
				}

				// Get the task state after first completion
				tasksAfterFirst := tl.ListTasks()
				var taskAfterFirst *models.Task
				for i := range tasksAfterFirst {
					if tasksAfterFirst[i].ID == selectedTask.ID {
						taskAfterFirst = &tasksAfterFirst[i]
						break
					}
				}

				if taskAfterFirst == nil || !taskAfterFirst.Completed {
					return false
				}

				// Complete the task multiple more times
				for i := 0; i < repeatCount; i++ {
					err = tl.CompleteTask(selectedTask.ID)
					if err != nil {
						return false // Should not error on already completed task
					}
				}

				// Get the task state after multiple completions
				tasksAfterMultiple := tl.ListTasks()
				var taskAfterMultiple *models.Task
				for i := range tasksAfterMultiple {
					if tasksAfterMultiple[i].ID == selectedTask.ID {
						taskAfterMultiple = &tasksAfterMultiple[i]
						break
					}
				}

				if taskAfterMultiple == nil {
					return false
				}

				// Verify the task is still completed (idempotent)
				if !taskAfterMultiple.Completed {
					return false
				}

				// Verify all other fields remain unchanged
				if taskAfterMultiple.ID != taskAfterFirst.ID {
					return false
				}
				if taskAfterMultiple.Description != taskAfterFirst.Description {
					return false
				}
				if !taskAfterMultiple.CreatedAt.Equal(taskAfterFirst.CreatedAt) {
					return false
				}

				// Verify the total number of tasks hasn't changed
				if len(tasksAfterMultiple) != len(addedTasks) {
					return false
				}

				return true
			},
			// Generate slices of non-empty strings (at least 1 task)
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			).SuchThat(func(s []string) bool {
				return len(s) >= 1
			}),
			// Generate a task index to select which task to complete
			gen.IntRange(0, 1000),
			// Generate number of times to repeat the complete operation (1-10 times)
			gen.IntRange(1, 10),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 10: 无效 ID 操作返回错误
// For any non-existent task ID, attempting to complete or delete that task should return an error and the task list should remain unchanged
// Validates: Requirements 3.2, 4.2
func TestProperty_InvalidIDOperationsReturnError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("operations with invalid IDs return errors and don't modify list",
		prop.ForAll(
			func(descriptions []string, invalidID int) bool {
				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks
				addedTasks := make([]*models.Task, 0, len(descriptions))
				for _, desc := range descriptions {
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
					addedTasks = append(addedTasks, task)
				}

				// Get initial state
				initialTasks := tl.ListTasks()
				initialCount := len(initialTasks)

				// Ensure invalidID doesn't exist in the list
				for _, task := range addedTasks {
					if task.ID == invalidID {
						// Skip this test case if the generated ID happens to exist
						return true
					}
				}

				// Test CompleteTask with invalid ID
				err = tl.CompleteTask(invalidID)

				// Verify appropriate error is returned
				if invalidID <= 0 {
					// Negative or zero IDs should return apperrors.ErrInvalidID
					if err != apperrors.ErrInvalidID {
						return false
					}
				} else {
					// Positive non-existent IDs should return apperrors.ErrTaskNotFound
					if err != apperrors.ErrTaskNotFound {
						return false
					}
				}

				// Verify list is unchanged after failed CompleteTask
				tasksAfterComplete := tl.ListTasks()
				if len(tasksAfterComplete) != initialCount {
					return false
				}

				// Verify all tasks remain unchanged
				for i, task := range tasksAfterComplete {
					if task.ID != initialTasks[i].ID {
						return false
					}
					if task.Description != initialTasks[i].Description {
						return false
					}
					if task.Completed != initialTasks[i].Completed {
						return false
					}
					if !task.CreatedAt.Equal(initialTasks[i].CreatedAt) {
						return false
					}
				}

				// Test DeleteTask with invalid ID
				err = tl.DeleteTask(invalidID)

				// Verify appropriate error is returned
				if invalidID <= 0 {
					// Negative or zero IDs should return apperrors.ErrInvalidID
					if err != apperrors.ErrInvalidID {
						return false
					}
				} else {
					// Positive non-existent IDs should return apperrors.ErrTaskNotFound
					if err != apperrors.ErrTaskNotFound {
						return false
					}
				}

				// Verify list is unchanged after failed DeleteTask
				tasksAfterDelete := tl.ListTasks()
				if len(tasksAfterDelete) != initialCount {
					return false
				}

				// Verify all tasks remain unchanged
				for i, task := range tasksAfterDelete {
					if task.ID != initialTasks[i].ID {
						return false
					}
					if task.Description != initialTasks[i].Description {
						return false
					}
					if task.Completed != initialTasks[i].Completed {
						return false
					}
					if !task.CreatedAt.Equal(initialTasks[i].CreatedAt) {
						return false
					}
				}

				return true
			},
			// Generate slices of non-empty strings (can be empty list too)
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			),
			// Generate invalid IDs: negative numbers, zero, and large positive numbers unlikely to exist
			gen.OneGenOf(
				gen.IntRange(-1000, 0),    // Negative and zero IDs
				gen.IntRange(1000, 10000), // Large positive IDs unlikely to exist
			),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: todo-list-cli, Property 11: 删除任务移除任务
// For any existing task ID, calling DeleteTask should remove the task from the list, and subsequent queries for that ID should fail
// Validates: Requirements 4.1
func TestProperty_DeleteTaskRemovesTask(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("deleting a task removes it from the list",
		prop.ForAll(
			func(descriptions []string, taskIndex int) bool {
				// Skip empty lists
				if len(descriptions) == 0 {
					return true
				}

				// Create fresh storage and todolist for each test
				storage := &mockStorage{data: nil}
				tl, err := NewTodoList(storage)
				if err != nil {
					return false
				}

				// Add all tasks
				addedTasks := make([]*models.Task, 0, len(descriptions))
				for _, desc := range descriptions {
					task, err := tl.AddTask(desc)
					if err != nil {
						return false
					}
					addedTasks = append(addedTasks, task)
				}

				// Get initial count
				initialCount := len(tl.ListTasks())

				// Select a task to delete (using modulo to ensure valid index)
				selectedTask := addedTasks[taskIndex%len(addedTasks)]
				deletedID := selectedTask.ID

				// Delete the task
				err = tl.DeleteTask(deletedID)
				if err != nil {
					return false
				}

				// Get the updated task list
				listedTasks := tl.ListTasks()

				// Verify the list is now one task shorter
				if len(listedTasks) != initialCount-1 {
					return false
				}

				// Verify the deleted task is not in the list
				for _, task := range listedTasks {
					if task.ID == deletedID {
						return false // Deleted task still in list!
					}
				}

				// Verify subsequent operations on the deleted ID fail
				err = tl.CompleteTask(deletedID)
				if err != apperrors.ErrTaskNotFound {
					return false // Should return apperrors.ErrTaskNotFound
				}

				// Verify deleting the same ID again also fails
				err = tl.DeleteTask(deletedID)
				if err != apperrors.ErrTaskNotFound {
					return false // Should return apperrors.ErrTaskNotFound
				}

				// Verify all other tasks are still present
				for _, originalTask := range addedTasks {
					if originalTask.ID == deletedID {
						continue // Skip the deleted task
					}

					// Find this task in the list
					found := false
					for _, listedTask := range listedTasks {
						if listedTask.ID == originalTask.ID {
							found = true
							// Verify all fields are preserved
							if listedTask.Description != originalTask.Description {
								return false
							}
							if listedTask.Completed != originalTask.Completed {
								return false
							}
							if !listedTask.CreatedAt.Equal(originalTask.CreatedAt) {
								return false
							}
							break
						}
					}

					if !found {
						return false // A non-deleted task is missing!
					}
				}

				return true
			},
			// Generate slices of non-empty strings (at least 1 task)
			gen.SliceOf(
				gen.AnyString().SuchThat(func(s string) bool {
					return strings.TrimSpace(s) != ""
				}),
			).SuchThat(func(s []string) bool {
				return len(s) >= 1
			}),
			// Generate a task index to select which task to delete
			gen.IntRange(0, 1000),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
