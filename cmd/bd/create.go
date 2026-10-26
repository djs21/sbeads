package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	var (
		title       string
		issueType   string
		priority    int
		description string
		blocked     int
	)

	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a new issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" && len(args) > 0 {
				title = args[0]
			}
			if title == "" {
				return errors.New("--title is required")
			}
			if priority < 0 {
				return errors.New("priority must be non-negative")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			issue, err := appCtx.Service.Create(title, issueType, priority, description, blocked)
			if err != nil {
				return fmt.Errorf("creating issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created issue: %s (%s)\n", issue.ID, issue.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&title, "title", "t", "", "Issue title")
	flags.IntVarP(&priority, "priority", "p", 0, "Priority (non-negative)")
	flags.StringVarP(&description, "description", "d", "", "Description")
	flags.StringVar(&issueType, "issue-type", "task", "Issue type")
	flags.StringVarP(&issueType, "type", "y", "task", "Issue type (alias)")
	flags.IntVarP(&blocked, "blocked", "b", 0, "Is blocked (0 or 1)")

	return cmd
}
