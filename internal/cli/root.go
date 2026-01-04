package cli

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewMSKCmd(service app.MSKService) *cobra.Command {
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
			ctx := cmd.Context()
			mk, err := ensurePassword()
			if err != nil {
				return err
			}

			service.ConfigMK(ctx, mk)

			return nil
		},
	}

	cmd.PersistentFlags().StringP("master", "m", "", "Set the master key manually.")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	passwordCmd := NewPasswordCmd(service)
	cmd.AddCommand(passwordCmd)

	return cmd
}


func ensurePassword() ([]byte, error) {
	fmt.Print("Enter master key: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	if err != nil {
		return nil, err
	}

	if len(bytePassword) == 0 {
		return nil, errors.New("Invalid master key.")
	}

	return bytePassword, nil
}


