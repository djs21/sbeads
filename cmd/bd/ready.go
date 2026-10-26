package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewReadyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "List all ready issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			issues, err := appCtx.Service.Ready()
			if err != nil {
				return fmt.Errorf("listing ready issues: %w", err)
			}

			if len(issues) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No ready issues")
				return nil
			}

			for _, issue := range issues {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", issue.ID, issue.Title)
			}
			return nil
		},
	}
}
