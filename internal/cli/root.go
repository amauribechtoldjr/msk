package cli

import (
	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/spf13/cobra"
)

func NewMSKCmd(service *app.MSKService) *cobra.Command {
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
	}

	addCmd := NewAddCmd(service)
	cmd.AddCommand(addCmd)

	getCmd := NewGetCmd(service)
	cmd.AddCommand(getCmd)

	delCmd := NewDeleteCmd(service)
	cmd.AddCommand(delCmd)

	listCmd := NewListCmd(service)
	cmd.AddCommand(listCmd)

	updateCmd := NewUpdateCmd(service)
	cmd.AddCommand(updateCmd)

	return cmd
}
