package cli

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/instance"
	"github.com/CDX-1/pocketry/internal/server"
	"github.com/spf13/cobra"
)

func newServerCommand(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Create and manage Pocketry servers",
	}

	cmd.AddCommand(
		newServerInitCommand(),
		newServerRunCommand(opts),
		newServerInfoCommand(opts),
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

func newServerRunCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run a Pocketry server",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			directory := opts.path

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

func newServerInfoCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show server information",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			directory := opts.path

			inst, err := instance.Load(directory)
			if err != nil {
				return fmt.Errorf("load Pocketry server: %w", err)
			}

			store, err := db.Open(inst.DatabasePath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer func() {
				_ = store.Close()
			}()

			userCount, err := store.Q.GetUserCount(cmd.Context())
			if err != nil {
				return fmt.Errorf("get user count: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Pocketry server information:")
			fmt.Fprintln(cmd.OutOrStdout(), "Created:", inst.Marker.CreatedAt.Format(time.RFC1123))
			fmt.Fprintln(cmd.OutOrStdout(), "Directory:", inst.RootDir)
			fmt.Fprintln(cmd.OutOrStdout(), "Listen address:", net.JoinHostPort(
				inst.Config.Server.Host,
				strconv.Itoa(inst.Config.Server.Port),
			))
			fmt.Fprintln(cmd.OutOrStdout(), "Config version:", inst.Config.Version)
			fmt.Fprintln(cmd.OutOrStdout(), "Instance format version:", inst.Marker.FormatVersion)
			fmt.Fprintln(cmd.OutOrStdout(), "Users:", userCount)

			return nil
		},
	}
}