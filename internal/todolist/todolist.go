package todolist

import (
	"strings"
	"time"
	apperrors "todolist/internal/errors"
	"todolist/internal/models"
	"todolist/internal/storage"
)

// TodoList manages the core business logic for todo items
type TodoList struct {
	list    *models.TaskList
	storage storage.Storage
}

// NewTodoList creates a new TodoList instance and loads initial data from storage
func NewTodoList(storage storage.Storage) (*TodoList, error) {
	list, err := storage.Load()
	if err != nil {
		return nil, apperrors.WrapWithContext(err, "failed to initialize todo list")
	}

	return &TodoList{
		list:    list,
		storage: storage,
	}, nil
}

// AddTask adds a new task to the list
func (tl *TodoList) AddTask(description string) (*models.Task, error) {
	// Validate description is not empty after trimming whitespace
	if strings.TrimSpace(description) == "" {
		return nil, apperrors.ErrEmptyDescription
	}

	// Create new task
	task := models.Task{
		ID:          tl.list.NextID,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	// Add to task list
	tl.list.Tasks = append(tl.list.Tasks, task)
	tl.list.NextID++

	// Save to storage
	if err := tl.storage.Save(tl.list); err != nil {
		// Rollback on save failure
		tl.list.Tasks = tl.list.Tasks[:len(tl.list.Tasks)-1]
		tl.list.NextID--
		return nil, apperrors.WrapWithContext(err, "failed to save task after adding")
	}

	return &task, nil
}

// ListTasks returns a copy of all tasks sorted by creation time
func (tl *TodoList) ListTasks() []models.Task {
	// Create a copy of the tasks slice
	tasks := make([]models.Task, len(tl.list.Tasks))
	copy(tasks, tl.list.Tasks)

	// Tasks are already sorted by creation time due to sequential addition
	// But we'll ensure it explicitly for correctness
	// Since IDs are sequential and CreatedAt is set on creation,
	// the natural order is already by creation time

	return tasks
}

// CompleteTask marks a task as completed
func (tl *TodoList) CompleteTask(id int) error {
	// Validate ID
	if id <= 0 {
		return apperrors.ErrInvalidID
	}

	// Find task by ID
	taskIndex := -1
	for i, task := range tl.list.Tasks {
		if task.ID == id {
			taskIndex = i
			break
		}
	}

	// Task not found
	if taskIndex == -1 {
		return apperrors.ErrTaskNotFound
	}

	// Mark as completed
	tl.list.Tasks[taskIndex].Completed = true

	// Save to storage
	if err := tl.storage.Save(tl.list); err != nil {
		// Rollback on save failure
		tl.list.Tasks[taskIndex].Completed = false
		return apperrors.WrapWithContext(err, "failed to save task after completing")
	}

	return nil
}

// DeleteTask removes a task from the list
func (tl *TodoList) DeleteTask(id int) error {
	// Validate ID
	if id <= 0 {
		return apperrors.ErrInvalidID
	}

	// Find task by ID
	taskIndex := -1
	for i, task := range tl.list.Tasks {
		if task.ID == id {
			taskIndex = i
			break
		}
	}

	// Task not found
	if taskIndex == -1 {
		return apperrors.ErrTaskNotFound
	}

	// Store deleted task for potential rollback
	deletedTask := tl.list.Tasks[taskIndex]

	// Remove task from list
	tl.list.Tasks = append(tl.list.Tasks[:taskIndex], tl.list.Tasks[taskIndex+1:]...)

	// Save to storage
	if err := tl.storage.Save(tl.list); err != nil {
		// Rollback on save failure - insert task back at original position
		tl.list.Tasks = append(tl.list.Tasks[:taskIndex], append([]models.Task{deletedTask}, tl.list.Tasks[taskIndex:]...)...)
		return apperrors.WrapWithContext(err, "failed to save task after deleting")
	}

	return nil
}
