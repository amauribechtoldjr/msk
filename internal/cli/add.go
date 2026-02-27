package cli

import (
	"errors"
	"fmt"

	clip "github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/generator"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewAddCmd(holder *ServiceHolder) *cobra.Command {
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

			name := args[0]

			err := validator.Validate(name)
			if err != nil {
				return fmt.Errorf("invalid password name: %v", err)
			}

			generate, _ := cmd.Flags().GetBool("generate")
			length, _ := cmd.Flags().GetInt("length")
			noSymbols, _ := cmd.Flags().GetBool("no-symbols")

			var password []byte

			if generate {
				password, err = generator.GeneratePassword(length, noSymbols)
				if err != nil {
					return fmt.Errorf("failed to generate password: %w", err)
				}
			} else {
				password, err = PromptSafeValue("Enter password:")
				if err != nil {
					return err
				}
			}
			defer wipe.Bytes(password)

			err = holder.Service.AddSecret(name, password)
			if err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}

			secret, err := holder.Service.GetSecret(name)
			if err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}
			defer wipe.Bytes(secret)

			if generate {
				err = clip.CopyText(secret)
				if err != nil {
					return fmt.Errorf("failed to copy password to your clipboard: %w", err)
				}

				logger.PrintSuccess("Password generated and copied to clipboard (press Ctrl+V to paste)\n\n")

				clip.Clear()
			} else {
				logger.PrintSuccess("Password added successfully\n")
			}

			return nil
		},
	}

	addCmd.Flags().BoolP("generate", "g", false, "Generate a random password instead of prompting")
	addCmd.Flags().IntP("length", "l", 16, "Length of the generated password")
	addCmd.Flags().Bool("no-symbols", false, "Exclude symbols from the generated password")

	return addCmd
}
