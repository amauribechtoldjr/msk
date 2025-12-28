package cli

import (
	"os"

	"github.com/spf13/cobra"
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
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


