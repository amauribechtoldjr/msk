package cli

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
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
		return ensurePassword()
	},
}

//TODO: 
// Review this later, maybe there are safer ways of using the masterKey, 
// I should probably hash it and clear from memory after usage
var masterKey string

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&masterKey, "master", "m", "", "Set the master key manually.")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ensurePassword() error {
	if masterKey != "" {
		return nil
	}

	fmt.Print("Enter master key: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	if err != nil {
		return err
	}

	if len(bytePassword) == 0 {
		return errors.New("Invalid master key.")
	}

	masterKey = string(bytePassword)
	return nil
}


