package cli

import (
	"errors"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(service *app.MSKService) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:           "update <name>",
		Aliases:       []string{"u"},
		Short:         "Used to update passwords of the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			mk, err := PromptMasterPassword(false)
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

			err := validator.Validate(name)
			if err != nil {
				return fmt.Errorf("invalid password name")
			}

			password, err := PromptSafeValue("Enter password:")
			if err != nil {
				return err
			}
			defer wipe.Bytes(password)

			err = service.UpdateSecret(name, password)
			if err != nil {
				return fmt.Errorf("failed to update secret: %w", err)
			}

			logger.PrintSuccess("Password updated successfully\n")
			return nil
		},
	}

	return updateCmd
}
