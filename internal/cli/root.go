package cli

import (
	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/spf13/cobra"
)

type ServiceHolder struct {
	Service *app.MSKService
}

func NewMSKCmd(enc encryption.Encryption) *cobra.Command {
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
			mk, err := PromptMasterPassword(false)
			if err != nil {
				return err
			}

			enc.ConfigMK(mk)

			vaultPath, err := config.Load(enc)
			if err != nil {
				enc.DestroyMK()
				return err
			}

			store, err := storage.NewStore(vaultPath)
			if err != nil {
				enc.DestroyMK()
				return err
			}

			holder.Service = app.NewMSKService(store, enc)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if holder.Service != nil {
				holder.Service.DestroyMK()
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

	configCmd := NewConfigCmd(enc)
	cmd.AddCommand(configCmd)

	return cmd
}
