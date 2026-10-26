package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"sbeads/internal/domain"
)

func NewUnclaimCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "unclaim [issue-id]",
		Short: "Unclaim an issue",
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

			issueDomain := domain.NewIssueDomain(appCtx.Storage)
			issue, err := issueDomain.Unclaim(id)
			if err != nil {
				return fmt.Errorf("unclaiming issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Unclaimed issue: %s (%s)\n", issue.ID, issue.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&id, "issue-id", "i", "", "Issue ID")

	return cmd
}
