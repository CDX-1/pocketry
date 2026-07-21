package cli

import (
	"fmt"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/CDX-1/pocketry/internal/admin"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/instance"
	"github.com/spf13/cobra"
)

func newUserCommand(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Fetch and manage registered users",
	}

	cmd.AddCommand(
		newUserListCommand(opts),
		newUserGetCommand(opts),
		newUserDeleteCommand(opts),
	)

	return cmd
}

func newUserListCommand(opts *rootOptions) *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered users",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			inst, err := instance.Load(opts.path)
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

			service := admin.NewService(store.Q)

			users, err := service.ListUsers(cmd.Context(), limit)
			if err != nil {
				return err
			}

			if len(users) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No registered users found.")
				return nil
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			defer writer.Flush()

			fmt.Fprintf(writer, "ID\tUsername\tCryptoPolicyVersion\tCreatedAt\n")
			fmt.Fprintf(writer, "---\t--------\t----------------\t---------\n")

			for _, user := range users {
				fmt.Fprintf(
					writer,
					"%d\t%s\t%d\t%s\n",
					user.ID,
					user.Username,
					user.CryptoPolicyVersion,
					user.CreatedAt.Format("2006-01-02 15:04:05"),
				)
			}

			return nil
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 100, "maximum numbers of users to list")

	return cmd
}

func newUserGetCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <user_id>",
		Short: "Get user details by user id",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid user ID: %w", err)
			}

			inst, err := instance.Load(opts.path)
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

			service := admin.NewService(store.Q)

			user, err := service.GetUser(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("get user: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "User details:")
			fmt.Fprintln(cmd.OutOrStdout(), "ID:", user.ID)
			fmt.Fprintln(cmd.OutOrStdout(), "Username:", user.Username)
			fmt.Fprintln(cmd.OutOrStdout(), "Username Normalized:", user.UsernameNormalized)
			fmt.Fprintln(cmd.OutOrStdout(), "Crypto Policy Version:", user.CryptoPolicyVersion)
			fmt.Fprintln(cmd.OutOrStdout(), "Created At:", user.CreatedAt.Format(time.RFC3339))
			fmt.Fprintln(cmd.OutOrStdout(), "Updated At:", user.UpdatedAt.Format(time.RFC3339))

			return nil
		},
	}
}

func newUserDeleteCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <user_id>",
		Short: "Delete a user",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid user ID: %w", err)
			}

			inst, err := instance.Load(opts.path)
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

			service := admin.NewService(store.Q)

			err = service.DeleteUser(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("delete user: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "User deleted successfully.")

			return nil
		},
	}
}