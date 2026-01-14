package main

import (
	"fmt"
	"os"
	"path/filepath"
	"todolist/internal/cli"
	"todolist/internal/storage"
	"todolist/internal/todolist"
)

func main() {
	// Get home directory for default storage path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize FileStorage with default path ~/.todolist.json
	storagePath := filepath.Join(homeDir, ".todolist.json")
	fileStorage := storage.NewFileStorage(storagePath)

	// Create TodoList instance
	tl, err := todolist.NewTodoList(fileStorage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize todo list: %v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments (skip program name)
	args := os.Args[1:]
	if len(args) == 0 {
		// No command provided, show help
		args = []string{"help"}
	}

	// Parse command
	cmd, err := cli.ParseCommand(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nUse 'todolist help' for usage information.")
		os.Exit(1)
	}

	// Execute command
	output, err := cli.ExecuteCommand(cmd, tl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Display result
	fmt.Println(output)
}
