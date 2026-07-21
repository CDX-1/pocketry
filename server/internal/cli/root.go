package cli

import (
	"fmt"

	"github.com/CDX-1/pocketry/internal/buildinfo"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	path string
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	opts := &rootOptions{}
	var showVersion bool

	cmd := &cobra.Command{
		Use:   "pocketry",
		Short: "Pocketry self-hosted password manager",

		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), buildinfo.VersionString())
				return nil
			}

			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(
		&opts.path,
		"path",
		"p",
		".",
		"path to the Pocketry server instance",
	)

	cmd.PersistentFlags().BoolVarP(
		&showVersion,
		"version",
		"v",
		false,
		"Print Pocketry build information",
	)

	cmd.AddCommand(
		newServerCommand(opts),
		newUsersCommand(opts),
		newVersionCommand(),
	)

	return cmd
}