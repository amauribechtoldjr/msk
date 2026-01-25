package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/spf13/cobra"
)

func NewListCmd(service app.MSKService) *cobra.Command {
	listCmd := &cobra.Command{
		Use:           "list <name>",
		Aliases:       []string{"l"},
		Short:         "Used to list passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			_, err := service.ListSecrets(ctx)
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			logger.PrintSuccess("Password copied to clipboard (press Ctrl+V to paste)")
			return nil
		},
	}

	return listCmd
}
