package cli

import "github.com/spf13/cobra"

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pocketry",
		Short: "Pocketry self-hosted password manager",
	}

	cmd.AddCommand(
		newServerCommand(),
	)
	
	return cmd
}