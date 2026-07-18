package cli

import (
	"fmt"

	"github.com/CDX-1/pocketry/internal/instance"
	"github.com/spf13/cobra"
)

func newServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Create and manage Pocketry servers",
	}

	cmd.AddCommand(
		newServerInitCommand(),
		newServerRunCommand(),
	)

	return cmd
}

func newServerInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <directory>",
		Short: "Create a new Pocketry server",
		Args:  cobra.ExactArgs(1),
		
		RunE: func(cmd *cobra.Command, args []string) error {
			directory := args[0]

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Creating Pocketry server in %s\n",
				directory,
			)

			inst, err := instance.Initialize(directory)
			if err != nil {
				return fmt.Errorf("initialize Pocketry server: %w", err)
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Pocketry server initialized successfully.\n\nDirectory: %s\n",
				inst.RootDir,
			)

			return nil
		},
	}
}

func newServerRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run <directory>",
		Short: "Run a Pocketry server",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			directory := args[0]

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Running Pocketry server from %s\n",
				directory,
			)

			return nil
		},
	}
}