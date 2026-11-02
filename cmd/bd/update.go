package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	var (
		id          string
		title       string
		issueType   string
		description string
		priority    int
		blocked     int
	)

	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update mutable fields of an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			if id == "" {
				return errors.New("issue ID is required")
			}

			flags := cmd.Flags()
			var pTitle, pType, pDesc *string
			var pPriority, pBlocked *int

			if flags.Changed("title") {
				pTitle = &title
			}
			if flags.Changed("type") || flags.Changed("issue-type") {
				pType = &issueType
			}
			if flags.Changed("description") {
				pDesc = &description
			}
			if flags.Changed("priority") {
				pPriority = &priority
			}
			if flags.Changed("blocked") {
				pBlocked = &blocked
			}

			if pTitle == nil && pType == nil && pDesc == nil && pPriority == nil && pBlocked == nil {
				return errors.New("at least one field flag (--title, --type, --description, --priority, --blocked) must be provided")
			}

			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Service == nil {
				return errors.New("service domain unavailable in context")
			}

			updated, err := appCtx.Service.Update(id, pTitle, pType, pDesc, pPriority, pBlocked)
			if err != nil {
				return fmt.Errorf("updating issue: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated issue: %s (%s)\n", updated.ID, updated.Title)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "Issue ID")
	flags.StringVarP(&title, "title", "t", "", "New issue title")
	flags.StringVarP(&issueType, "type", "y", "", "New issue type (bug, feature, task, epic)")
	flags.StringVar(&issueType, "issue-type", "", "New issue type (alias)")
	flags.StringVarP(&description, "description", "d", "", "New issue description")
	flags.IntVarP(&priority, "priority", "p", 0, "New priority rating")
	flags.IntVarP(&blocked, "blocked", "b", 0, "Set blocked state (0 or 1)")

	return cmd
}
