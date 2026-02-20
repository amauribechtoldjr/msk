package cli

import (
	"errors"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	clip "github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewGetCmd(service *app.MSKService) *cobra.Command {
	getCmd := &cobra.Command{
		Use:           "get <name>",
		Aliases:       []string{"g"},
		Short:         "Used to get passwords from the vault.",
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

			if err := validator.Validate(name); err != nil {
				return fmt.Errorf("invalid password name: %w", err)
			}

			password, err := service.GetSecret(name)
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}
			defer wipe.Bytes(password)

			err = clip.CopyText(password)
			if err != nil {
				wipe.Bytes(password)
				return fmt.Errorf("failed to copy password to your clipboard: %w", err)
			}

			logger.PrintSuccess("Password copied to clipboard (press Ctrl+V to paste)\n\n")

			clip.Clear()

			return nil
		},
	}

	return getCmd
}
