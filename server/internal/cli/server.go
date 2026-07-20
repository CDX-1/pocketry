package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CDX-1/pocketry/internal/instance"
	"github.com/CDX-1/pocketry/internal/server"
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

			inst, err := instance.Load(directory)
			if err != nil {
				return fmt.Errorf("load Pocketry server: %w", err)
			}

			ctx, stop := signal.NotifyContext(
				cmd.Context(),
				os.Interrupt,
				syscall.SIGTERM,
			)
			defer stop()

			return server.Run(ctx, cmd.OutOrStdout(), inst)
		},
	}
}