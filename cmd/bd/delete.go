package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeleteCmd() *cobra.Command {
	var (
		id      string
		cascade bool
	)

	cmd := &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete an issue",
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

			if err := appCtx.Service.Delete(id, cascade); err != nil {
				return fmt.Errorf("deleting issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted issue: %s\n", id)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")
	flags.BoolVarP(&cascade, "force", "f", false, "Force delete and cascade remove all dependencies")
	flags.BoolVarP(&cascade, "cascade", "c", false, "Cascade remove all dependencies")

	return cmd
}
