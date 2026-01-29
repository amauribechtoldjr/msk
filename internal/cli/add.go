package cli

import (
	"errors"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/spf13/cobra"
)

func NewAddCmd(service app.MSKService) *cobra.Command {
	addCmd := &cobra.Command{
		Use:           "add <name>",
		Aliases:       []string{"a"},
		Short:         "Used to add passwords to the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("password name is required")
			}

			ctx := cmd.Context()
			name := args[0]

			err := validator.Validate(name)
			if err != nil {
				return fmt.Errorf("invalid password name: %w", err)
			}

			value, err := cmd.Flags().GetString("password")
			if err != nil {
				return fmt.Errorf("failed to retrieve password value: %w", err)
			}

			password := []byte(value)

			if value == "" {
				var err error
				password, err = PromptPassword("Enter password:")
				if err != nil {
					return err
				}
			}

			err = service.AddSecret(ctx, name, password)
			if err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}

			logger.PrintSuccess("Password added successfully\n")
			return nil
		},
	}

	addCmd.Flags().StringP("password", "p", "", "Password identifier.")

	return addCmd
}
