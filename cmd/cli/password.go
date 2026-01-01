package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/file_manager"
	"github.com/amauribechtoldjr/msk/utils"
	"github.com/spf13/cobra"
)

var passwordCmd = &cobra.Command{
	Use:   "password",
	Aliases: []string{"p"},
	Short: "Used to add and get passwords from the MSK.",
	Long: ``,
	Run: password,
}

func password(cmd *cobra.Command, args []string) {
	mk, _ := cmd.Flags().GetString("master")
	pName, _ := cmd.Flags().GetString("name")

	shouldDelete, _ := cmd.Flags().GetBool("delete")

	if shouldDelete {
		err := file_manager.DeletePassword([]byte(mk), pName)
		if  err != nil {
			utils.ErrorMessage("Failed to delete password")
		}

		utils.SuccessMessage("Password deleted successfully")
		return
	}

	//TODO: move this to require at runtime level with prompt
	pValue, _ := cmd.Flags().GetString("new")

	if pValue != "" {
		err := file_manager.AddPassword([]byte(mk), pName, pValue)
		if  err != nil {
			err := fmt.Errorf("Failed to add password: %w", err)
			utils.ErrorMessage(err.Error())
			return
		}

		utils.SuccessMessage("Password added successfully")
		return
	}
	

	shouldListAll, _ := cmd.Flags().GetBool("list")

	if shouldListAll {
		err := file_manager.ListAll([]byte(mk))
		if  err != nil {
			utils.ErrorMessage("Failed to list passwords")
			return
		}
	}
}

func init() {
	rootCmd.AddCommand(passwordCmd)

	passwordCmd.Flags().StringP("name", "n", "", "Password identifier.")
	passwordCmd.Flags().StringP("new", "s", "", "Password value.")
	passwordCmd.Flags().BoolP("delete", "d", false, "Delete a password.")
	passwordCmd.Flags().BoolP("list", "l", false, "List all passwords.")
}
