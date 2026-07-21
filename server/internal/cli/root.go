package cli

import "github.com/spf13/cobra"

type rootOptions struct {
	path string
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   "pocketry",
		Short: "Pocketry self-hosted password manager",
	}

	cmd.PersistentFlags().StringVarP(
		&opts.path,
		"path",
		"p",
		".",
		"path to the Pocketry server instance",
	)

	cmd.AddCommand(
		newServerCommand(opts),
		newUsersCommand(opts),
	)

	return cmd
}