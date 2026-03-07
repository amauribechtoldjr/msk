package cli

import (
	"os"
	"slices"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/spf13/cobra"
)

type ServiceHolder struct {
	Service app.Service
}

var ignored_commands = []string{"version", "v", "help", "unlock", "lock", "config"}

func NewMSKCmd() *cobra.Command {
	holder := &ServiceHolder{}
	v := vault.NewMSKVault()

	cmd := &cobra.Command{
		Use:   "msk",
		Short: "MSK is a lightweight, offline password manager that securely encrypts your credentials using a master password.",
		Long: `
			MSK is a lightweight password manager designed to keep
			all your credentials securely stored on your own computer,
			without ever exposing them to the internet.
			All passwords are encrypted using a master password,
			ensuring that even if someone gains access to your machine,
			they won't be able to view any stored data without the correct master key.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if slices.Contains(ignored_commands, cmd.Name()) {
				return nil
			}

			helpFlag, err := cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}

			if helpFlag {
				return nil
			}

			holder.Service, err = app.BootstrapWithAuth(v)
			if err != nil {
				return err
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			v.DestroyMK()

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			isVersionCommand, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}

			if isVersionCommand {
				logger.PrintInfo(meta.Version)
				os.Exit(0)
			}

			return nil
		},
	}

	addCmd := NewAddCmd(holder)
	cmd.AddCommand(addCmd)

	getCmd := NewGetCmd(holder)
	cmd.AddCommand(getCmd)

	delCmd := NewDeleteCmd(holder)
	cmd.AddCommand(delCmd)

	listCmd := NewListCmd(holder)
	cmd.AddCommand(listCmd)

	updateCmd := NewUpdateCmd(holder)
	cmd.AddCommand(updateCmd)

	configCmd := NewConfigCmd(v)
	cmd.AddCommand(configCmd)

	versionCmd := NewVersionCmd()
	cmd.AddCommand(versionCmd)

	unlockCmd := NewUnlockCmd(v)
	cmd.AddCommand(unlockCmd)

	lockCmd := NewLockCmd()
	cmd.AddCommand(lockCmd)

	cmd.Flags().BoolP("version", "v", false, "Show MSK current version")

	return cmd
}
