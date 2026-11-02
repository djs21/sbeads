package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export issues and dependencies to a merge-friendly JSONL file",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Storage == nil {
				return fmt.Errorf("storage unavailable in context")
			}

			targetPath := outputFile
			if targetPath == "" {
				targetPath = filepath.Join(filepath.Dir(appCtx.DBPath), "issues.jsonl")
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("creating directory for export: %w", err)
			}

			f, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("creating export file: %w", err)
			}
			defer f.Close()

			if err := appCtx.Storage.ExportJSONL(f); err != nil {
				return fmt.Errorf("exporting data: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Exported issues and dependencies to %s\n", targetPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Target JSONL file path (defaults to .sbeads/issues.jsonl)")

	return cmd
}
