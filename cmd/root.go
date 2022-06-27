package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const Name = "discover"

// NewRoot returns a root cobra.Command for the whole CLI.
func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          Name,
		Short:        "Insights about a GitHub repository.",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(
		NewEngagement(),
	)

	return rootCmd
}
