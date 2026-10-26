package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage issue dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command: %s", args[0])
			}
			_ = cmd.Usage()
			return errors.New("subcommand required")
		},
	}

	cmd.AddCommand(newDepAddCmd())
	cmd.AddCommand(newDepListCmd())

	return cmd
}

func newDepAddCmd() *cobra.Command {
	var (
		childID  string
		parentID string
	)

	cmd := &cobra.Command{
		Use:   "add [child] [parent]",
		Short: "Add a dependency between issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			if childID == "" && len(args) >= 2 {
				childID = args[0]
				parentID = args[1]
			}
			if childID == "" || parentID == "" {
				return errors.New("--child and --parent are required")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			if err := appCtx.Service.AddDependency(childID, parentID); err != nil {
				return fmt.Errorf("adding dependency: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added dependency: %s -> %s\n", childID, parentID)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&childID, "child", "", "Child issue ID")
	flags.StringVar(&parentID, "parent", "", "Parent issue ID")

	return cmd
}

func newDepListCmd() *cobra.Command {
	var issueID string

	cmd := &cobra.Command{
		Use:   "list [issue-id]",
		Short: "List dependencies for an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if issueID == "" && len(args) > 0 {
				issueID = args[0]
			}
			if issueID == "" {
				return errors.New("--issue-id is required")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			deps, err := appCtx.Service.GetDependencies(issueID)
			if err != nil {
				return fmt.Errorf("listing dependencies: %w", err)
			}

			if len(deps) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No dependencies")
				return nil
			}

			for _, dep := range deps {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", dep)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&issueID, "issue-id", "", "Issue ID")

	return cmd
}
