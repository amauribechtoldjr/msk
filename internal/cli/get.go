package cli

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/app"
	clip "github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/spf13/cobra"
)

func NewGetCmd(service app.MSKService) *cobra.Command {
	getCmd := &cobra.Command{
		Use:           "get <name>",
		Aliases:       []string{"g"},
		Short:         "Used to get passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("password name is required")
			}

			ctx := cmd.Context()
			name := args[0]
	
			password, err := service.GetSecret(ctx, name)
			if err != nil {
				return err
			}

			err = clip.CopyText(password)
			if err != nil {
				return err
			}

			logger.PrintSuccess("Password copied to clipboard (press Ctrl+V to paste)")
			return nil
		},
	}

	return getCmd
}