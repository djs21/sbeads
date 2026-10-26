package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewClaimCmd() *cobra.Command {
	var (
		id       string
		assignee string
		agent    string
	)

	cmd := &cobra.Command{
		Use:   "claim [issue-id]",
		Short: "Claim an issue",
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

			issue, err := appCtx.Service.Claim(id, assignee, agent)
			if err != nil {
				return fmt.Errorf("claiming issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Claimed issue: %s (%s) by %s\n", issue.ID, issue.Title, agent)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&id, "issue-id", "i", "", "Issue ID")
	flags.StringVarP(&assignee, "assignee", "a", "", "Assignee")
	flags.StringVarP(&agent, "agent", "m", "me", "Agent name")

	return cmd
}
