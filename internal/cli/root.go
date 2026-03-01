package cli

import (
	"os"
	"slices"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/build"
	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/session"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

type ServiceHolder struct {
	Service *app.MSKService
}

var ignored_commands = []string{"version", "v", "help", "unlock", "lock", "config"}

func NewMSKCmd(vault vault.Vault) *cobra.Command {
	holder := &ServiceHolder{}

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

			// Session-unlock fast path
			if token := os.Getenv("MSK_SESSION"); token != "" {
				sess, err := session.New()
				if err == nil && sess.IsActive(token) {
					mk, err := sess.Load(token)
					if err == nil {
						defer wipe.Bytes(mk)
						vault.ConfigMK(mk)
						_ = sess.Refresh()

						vaultPath, err := config.Load(vault)
						if err != nil {
							vault.DestroyMK()
							return err
						}

						store, err := storage.NewStore(vaultPath)
						if err != nil {
							vault.DestroyMK()
							return err
						}

						holder.Service = app.NewMSKService(store, vault)
						return nil
					}
					_ = sess.Destroy()
				}
			}

			mk, err := PromptMasterPassword(false)
			if err != nil {
				return err
			}

			vault.ConfigMK(mk)

			vaultPath, err := config.Load(vault)
			if err != nil {
				vault.DestroyMK()
				return err
			}

			store, err := storage.NewStore(vaultPath)
			if err != nil {
				vault.DestroyMK()
				return err
			}

			holder.Service = app.NewMSKService(store, vault)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if holder.Service != nil {
				holder.Service.DestroyMK()
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			isVersionCommand, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}

			if isVersionCommand {
				logger.PrintInfo(build.Version)
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

	configCmd := NewConfigCmd(vault)
	cmd.AddCommand(configCmd)

	versionCmd := NewVersionCmd()
	cmd.AddCommand(versionCmd)

	unlockCmd := NewUnlockCmd(vault)
	cmd.AddCommand(unlockCmd)

	lockCmd := NewLockCmd()
	cmd.AddCommand(lockCmd)

	cmd.Flags().BoolP("version", "v", false, "Show MSK current version")

	return cmd
}
