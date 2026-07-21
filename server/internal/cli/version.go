package cli

import (
	"fmt"

	"github.com/CDX-1/pocketry/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Pocketry build information",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), buildinfo.VersionString())
			return nil
		},
	}
}