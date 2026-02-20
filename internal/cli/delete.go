package cli

import (
	"errors"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(service *app.MSKService) *cobra.Command {
	delCmd := &cobra.Command{
		Use:           "del <name>",
		Aliases:       []string{"d"},
		Short:         "Used to delete passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			mk, err := PromptMasterPassword(true)
			if err != nil {
				return err
			}

			service.ConfigMK(mk)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			service.DestroyMK()

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("password name is required")
			}

			name := args[0]

			if err := validator.Validate(name); err != nil {
				return fmt.Errorf("invalid password name: %w", err)
			}

			// I should be able to decrypt file with the master key first!!!
			// here its just deleting for now... (this is not safe)
			err := service.DeleteSecret(name)
			if err != nil {
				return err
			}

			logger.PrintSuccess("Password deleted successfully")
			return nil
		},
	}

	return delCmd
}
