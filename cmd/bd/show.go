package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sbeads/internal/domain"
	"sbeads/internal/types"
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

			edges, err := appCtx.Storage.GetDependencyEdges(issue.ID)
			if err != nil {
				return fmt.Errorf("fetching dependencies: %w", err)
			}

			var unclosedBlockers []types.Issue
			for _, b := range edges.Blockers {
				if b.IsClosed == 0 {
					unclosedBlockers = append(unclosedBlockers, b)
				}
			}

			status := domain.ComputeStatus(issue, len(unclosedBlockers) > 0)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Issue: %s\n", issue.ID)
			fmt.Fprintf(out, "Title: %s\n", issue.Title)
			fmt.Fprintf(out, "Type: %s\n", issue.IssueType)
			fmt.Fprintf(out, "Priority: %d\n", issue.Priority)
			fmt.Fprintf(out, "Status: %s %s\n", status, status.ShortBadge())

			agent := issue.Agent
			if agent == "" {
				agent = issue.Assignee
			}
			if (issue.IsClaimed != 0 || agent != "") && agent != "" {
				agent = strings.TrimPrefix(agent, "@")
				fmt.Fprintf(out, "Agent: @%s\n", agent)
			}

			fmt.Fprintf(out, "Description: %s\n", issue.Description)

			if len(unclosedBlockers) > 0 {
				fmt.Fprintln(out, "Blocked by:")
				for _, b := range unclosedBlockers {
					if b.Title != "" {
						fmt.Fprintf(out, "  %s: %s\n", b.ID, b.Title)
					} else {
						fmt.Fprintf(out, "  %s\n", b.ID)
					}
				}
			}

			if len(edges.Dependents) > 0 {
				fmt.Fprintln(out, "Blocks:")
				for _, d := range edges.Dependents {
					if d.Title != "" {
						fmt.Fprintf(out, "  %s: %s\n", d.ID, d.Title)
					} else {
						fmt.Fprintf(out, "  %s\n", d.ID)
					}
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&id, "issue-id", "i", "", "Issue ID")

	return cmd
}
