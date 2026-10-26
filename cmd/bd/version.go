package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print sbeads version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "sbeads version %s\n", Version)
		},
	}
}
