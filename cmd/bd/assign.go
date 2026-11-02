package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewAssignCmd() *cobra.Command {
	var (
		id       string
		assignee string
		agent    string
	)

	cmd := &cobra.Command{
		Use:   "assign [id]",
		Short: "Assign an issue to a user or agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				return errors.New("issue ID is required")
			}

			flags := cmd.Flags()
			if !flags.Changed("assignee") && !flags.Changed("agent") {
				return errors.New("at least one of --assignee or --agent must be provided")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Storage == nil {
				return errors.New("storage unavailable in context")
			}

			issue, err := appCtx.Storage.GetIssue(id)
			if err != nil {
				return fmt.Errorf("issue not found: %s", id)
			}

			if flags.Changed("assignee") {
				issue.Assignee = assignee
			}
			if flags.Changed("agent") {
				issue.Agent = agent
			}

			if err := appCtx.Storage.UpdateIssue(issue); err != nil {
				return fmt.Errorf("assigning issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Assigned issue: %s (%s)\n", issue.ID, issue.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")
	flags.StringVarP(&assignee, "assignee", "a", "", "Assignee user or role")
	flags.StringVarP(&agent, "agent", "m", "", "Agent identifier")

	return cmd
}
