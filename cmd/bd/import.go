package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewImportCmd() *cobra.Command {
	var inputFile string

	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"sync"},
		Short:   "Import/sync issues and dependencies from a JSONL file",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Storage == nil {
				return fmt.Errorf("storage unavailable in context")
			}

			sourcePath := inputFile
			if sourcePath == "" {
				sourcePath = filepath.Join(filepath.Dir(appCtx.DBPath), "issues.jsonl")
			}

			f, err := os.Open(sourcePath)
			if err != nil {
				return fmt.Errorf("opening import file: %w", err)
			}
			defer f.Close()

			issuesCount, depsCount, err := appCtx.Storage.ImportJSONL(f)
			if err != nil {
				return fmt.Errorf("importing data: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d issues and %d dependencies from %s\n", issuesCount, depsCount, sourcePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Source JSONL file path (defaults to .sbeads/issues.jsonl)")

	return cmd
}
