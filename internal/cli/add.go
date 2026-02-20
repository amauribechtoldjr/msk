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

func NewAddCmd(service *app.MSKService) *cobra.Command {
	addCmd := &cobra.Command{
		Use:           "add <name>",
		Aliases:       []string{"a"},
		Short:         "Used to add passwords to the vault.",
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

			err = service.AddSecret(name, password)
			if err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}

			logger.PrintSuccess("Password added successfully\n")
			return nil
		},
	}

	return addCmd
}
