package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCloseCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "close [issue-id]",
		Short: "Close an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				return errors.New("--issue-id is required")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			issue, err := appCtx.Service.Close(id)
			if err != nil {
				return fmt.Errorf("closing issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Closed issue: %s (%s)\n", issue.ID, issue.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&id, "issue-id", "i", "", "Issue ID")

	return cmd
}
