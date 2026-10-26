package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"sbeads/backend"
	"sbeads/internal/domain"
)

func NewRootCmd() *cobra.Command {
	var dbPathFlag string

	rootCmd := &cobra.Command{
		Use:           "sbeads",
		Short:         "Minimal issue tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "sbeads" || cmd.Name() == "version" || cmd.Name() == "help" {
				return nil
			}

			dbPath := dbPathFlag
			if dbPath == "" {
				dbPath = getDBPath()
			}

			if err := backend.CreateTables(dbPath); err != nil {
				return fmt.Errorf("creating tables: %w", err)
			}

			storage, err := backend.NewStorage(dbPath)
			if err != nil {
				return fmt.Errorf("creating storage: %w", err)
			}

			service := domain.NewServiceDomain(storage)
			appCtx := &AppContext{
				DBPath:  dbPath,
				Storage: storage,
				Service: service,
			}

			cmd.SetContext(WithAppContext(cmd.Context(), appCtx))
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			appCtx := GetAppContext(cmd.Context())
			if appCtx != nil && appCtx.Storage != nil {
				return appCtx.Storage.Close()
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("sbeads - Minimal issue tracker")
			cmd.Println("Usage: sbeads <command> [options]")
			cmd.Println("Commands: create, ready, claim, unclaim, close, show, dep, version")
			return errors.New("command required")
		},
	}

	rootCmd.PersistentFlags().StringVar(&dbPathFlag, "db", "", "Database path")

	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewReadyCmd())
	rootCmd.AddCommand(NewClaimCmd())
	rootCmd.AddCommand(NewUnclaimCmd())
	rootCmd.AddCommand(NewCloseCmd())
	rootCmd.AddCommand(NewShowCmd())
	rootCmd.AddCommand(NewDepCmd())
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}
