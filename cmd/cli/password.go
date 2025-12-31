package cli

import (
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

	//TODO: move this to require at execution level with prompt
	pValue, _ := cmd.Flags().GetString("new")

	err := file_manager.AddPassword([]byte(mk), pName, pValue)
	if  err != nil {
		utils.ErrorMessage("Failed to add password")
		return
	}
	
	utils.SuccessMessage("Password added successfully")
}

func init() {
	rootCmd.AddCommand(passwordCmd)

	passwordCmd.Flags().StringP("name", "n", "", "Password identifier.")
	passwordCmd.MarkFlagRequired("name")

	passwordCmd.Flags().StringP("new", "s", "", "Password value.")
	passwordCmd.Flags().BoolP("delete", "d", false, "Delete a password.")
}
