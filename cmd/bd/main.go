package main

import (
	"fmt"
	"os"
)

func Execute() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}

func ExecuteArgs(args []string) error {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func execute(args []string) error {
	if len(args) <= 1 {
		return ExecuteArgs([]string{})
	}
	return ExecuteArgs(args[1:])
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
