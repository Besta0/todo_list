package cli

import (
	"fmt"
	"strconv"
	"strings"
	apperrors "todolist/internal/errors"
	"todolist/internal/todolist"
)

// Command represents a parsed CLI command
type Command struct {
	Name string
	Args []string
}

// ParseCommand parses command line arguments into a Command structure
func ParseCommand(args []string) (*Command, error) {
	// Need at least one argument (the command name)
	if len(args) == 0 {
		return nil, apperrors.ErrInvalidCommand
	}

	cmdName := strings.ToLower(args[0])

	// Validate command name
	switch cmdName {
	case "add":
		// add command requires at least one argument (description)
		if len(args) < 2 {
			return nil, apperrors.WrapCommandError(apperrors.ErrInvalidCommand, "add command requires a description")
		}
		// Join all remaining args as the description
		description := strings.Join(args[1:], " ")
		return &Command{
			Name: "add",
			Args: []string{description},
		}, nil

	case "list":
		// list command takes no arguments
		return &Command{
			Name: "list",
			Args: []string{},
		}, nil

	case "done":
		// done command requires exactly one argument (task ID)
		if len(args) != 2 {
			return nil, apperrors.WrapCommandError(apperrors.ErrInvalidCommand, "done command requires a task ID")
		}
		// Validate that the argument is a valid integer
		if _, err := strconv.Atoi(args[1]); err != nil {
			return nil, apperrors.WrapCommandError(apperrors.ErrInvalidCommand, "task ID must be a valid number")
		}
		return &Command{
			Name: "done",
			Args: []string{args[1]},
		}, nil

	case "delete":
		// delete command requires exactly one argument (task ID)
		if len(args) != 2 {
			return nil, apperrors.WrapCommandError(apperrors.ErrInvalidCommand, "delete command requires a task ID")
		}
		// Validate that the argument is a valid integer
		if _, err := strconv.Atoi(args[1]); err != nil {
			return nil, apperrors.WrapCommandError(apperrors.ErrInvalidCommand, "task ID must be a valid number")
		}
		return &Command{
			Name: "delete",
			Args: []string{args[1]},
		}, nil

	case "help":
		// help command takes no arguments
		return &Command{
			Name: "help",
			Args: []string{},
		}, nil

	default:
		return nil, apperrors.ErrInvalidCommand
	}
}

// ExecuteCommand executes a parsed command and returns formatted output
func ExecuteCommand(cmd *Command, tl *todolist.TodoList) (string, error) {
	switch cmd.Name {
	case "add":
		// Add a new task
		task, err := tl.AddTask(cmd.Args[0])
		if err != nil {
			return "", apperrors.WrapCommandError(err, "add")
		}
		return fmt.Sprintf("✓ Task added: [%d] %s", task.ID, task.Description), nil

	case "list":
		// List all tasks
		tasks := tl.ListTasks()
		if len(tasks) == 0 {
			return "No tasks found. Add a task with: todolist add <description>", nil
		}

		var output strings.Builder
		output.WriteString("Your tasks:\n")
		for _, task := range tasks {
			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}
			output.WriteString(fmt.Sprintf("%s [%d] %s (created: %s)\n",
				status,
				task.ID,
				task.Description,
				task.CreatedAt.Format("2006-01-02 15:04:05")))
		}
		return strings.TrimSpace(output.String()), nil

	case "done":
		// Mark task as completed
		id, _ := strconv.Atoi(cmd.Args[0]) // Already validated in ParseCommand
		if err := tl.CompleteTask(id); err != nil {
			return "", apperrors.WrapCommandError(err, "done")
		}
		return fmt.Sprintf("✓ Task %d marked as completed", id), nil

	case "delete":
		// Delete a task
		id, _ := strconv.Atoi(cmd.Args[0]) // Already validated in ParseCommand
		if err := tl.DeleteTask(id); err != nil {
			return "", apperrors.WrapCommandError(err, "delete")
		}
		return fmt.Sprintf("✓ Task %d deleted", id), nil

	case "help":
		// Display help information
		return getHelpText(), nil

	default:
		return "", apperrors.ErrInvalidCommand
	}
}

// getHelpText returns the help message
func getHelpText() string {
	return `Todo List CLI - A simple command-line todo list manager

Usage:
  todolist <command> [arguments]

Commands:
  add <description>    Add a new task
  list                 List all tasks
  done <id>            Mark a task as completed
  delete <id>          Delete a task
  help                 Show this help message

Examples:
  todolist add "Buy groceries"
  todolist list
  todolist done 1
  todolist delete 2`
}
