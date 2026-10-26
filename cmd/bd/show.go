package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewShowCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "show [issue-id]",
		Short: "Show issue details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				return errors.New("--issue-id is required")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Storage == nil {
				return errors.New("storage unavailable in context")
			}

			issue, err := appCtx.Storage.GetIssue(id)
			if err != nil {
				return fmt.Errorf("issue not found: %s", id)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Issue: %s\n", issue.ID)
			fmt.Fprintf(out, "Title: %s\n", issue.Title)
			fmt.Fprintf(out, "Type: %s\n", issue.IssueType)
			fmt.Fprintf(out, "Priority: %d\n", issue.Priority)
			fmt.Fprintf(out, "Description: %s\n", issue.Description)
			fmt.Fprintf(out, "Blocked: %v\n", issue.IsBlocked)
			fmt.Fprintf(out, "Closed: %v\n", issue.IsClosed)
			fmt.Fprintf(out, "Claimed: %v\n", issue.IsClaimed)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&id, "issue-id", "i", "", "Issue ID")

	return cmd
}
