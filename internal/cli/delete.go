package cli

import (
	"errors"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(service app.MSKService) *cobra.Command {
	getCmd := &cobra.Command{
		Use:           "del <name>",
		Aliases:       []string{"d"},
		Short:         "Used to delete passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("password name is required")
			}

			ctx := cmd.Context()
			name := args[0]

			if err := validator.Validate(name); err != nil {
				return fmt.Errorf("invalid password name: %w", err)
			}

			err := service.DeleteSecret(ctx, name)
			if err != nil {
				return err
			}

			logger.PrintSuccess("Password deleted successfully")
			return nil
		},
	}

	return getCmd
}