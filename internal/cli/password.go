package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/utils"
	"github.com/spf13/cobra"
)


func NewPasswordCmd(service app.MSKService) *cobra.Command {
	passwordCmd := &cobra.Command{
		Use:   "password",
		Aliases: []string{"p"},
		Short: "Used to add and get passwords from the MSK.",
		Long: ``,
		RunE: func (cmd *cobra.Command, args []string) error {
			// mk, _ := cmd.Flags().GetString("master")
			pName, _ := cmd.Flags().GetString("name")

			ctx := cmd.Context()

			// shouldDelete, _ := cmd.Flags().GetBool("delete")

			// if shouldDelete {
			// 	err := file_manager.DeletePassword([]byte(mk), pName)
			// 	if  err != nil {
			// 		return fmt.Errorf("failed to delete password: %w", err)
			// 	}

			// 	return nil
			// }

			//TODO: move this to require at runtime level with prompt
			pValue, _ := cmd.Flags().GetString("new")

			if pValue != "" {
				err := service.AddSecret(ctx, pName, pValue)
				if  err != nil {
					return fmt.Errorf("failed to add password: %w", err)
				}

				utils.SuccessMessage("Password added successfully")
				return nil
			}
			
			// shouldListAll, _ := cmd.Flags().GetBool("list")

			// if shouldListAll {
			// 	err := file_manager.ListAll([]byte(mk))
			// 	if  err != nil {
			// 		return fmt.Errorf("failed to list passwords: %w", err)
			// 	}
			// }

			shouldGetSecret, _ := cmd.Flags().GetBool("get")

			if shouldGetSecret {
				err := service.GetSecret(ctx, pName)
				if  err != nil {
					return fmt.Errorf("failed to get passwords: %w", err)
				}
			}

			return nil
		},
	}

	passwordCmd.Flags().StringP("name", "n", "", "Password identifier.")
	passwordCmd.Flags().StringP("new", "s", "", "Password value.")
	passwordCmd.Flags().BoolP("delete", "d", false, "Delete a password.")
	passwordCmd.Flags().BoolP("list", "l", false, "List all passwords.")
	passwordCmd.Flags().BoolP("get", "g", false, "Get one passwords.")

	return passwordCmd
}
