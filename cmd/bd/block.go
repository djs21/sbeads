package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewBlockCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "block [id]",
		Short: "Block an issue manually",
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

			blocked := 1
			updated, err := appCtx.Service.Update(id, nil, nil, nil, nil, &blocked)
			if err != nil {
				return fmt.Errorf("blocking issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Blocked issue: %s (%s)\n", updated.ID, updated.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")

	return cmd
}

func NewUnblockCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "unblock [id]",
		Short: "Unblock a manually blocked issue",
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

			unblocked := 0
			updated, err := appCtx.Service.Update(id, nil, nil, nil, nil, &unblocked)
			if err != nil {
				return fmt.Errorf("unblocking issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Unblocked issue: %s (%s)\n", updated.ID, updated.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")

	return cmd
}
