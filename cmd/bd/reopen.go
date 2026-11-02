package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewReopenCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "reopen [id]",
		Short: "Reopen a closed issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				return errors.New("issue ID is required")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			reopened, err := appCtx.Service.Reopen(id)
			if err != nil {
				return fmt.Errorf("reopening issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Reopened issue: %s (%s)\n", reopened.ID, reopened.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")

	return cmd
}
