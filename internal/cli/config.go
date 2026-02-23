package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/spf13/cobra"
)

func NewConfigCmd(enc encryption.Encryption) *cobra.Command {
	configCmd := &cobra.Command{
		Use:           "config",
		Short:         "Configure MSK vault path and master password.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			enc.DestroyMK()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			exists, err := config.Exists()
			if err != nil {
				return fmt.Errorf("failed to check config: %w", err)
			}

			if exists {
				logger.PrintInfo("Config already exists. Overwrite? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					logger.PrintSuccess("Config unchanged.\n")
					return nil
				}
			}

			defaultPath, err := config.DefaultVaultPath()
			if err != nil {
				return fmt.Errorf("failed to get default vault path: %w", err)
			}

			fmt.Printf("Enter vault path (press Enter for default: %s): ", defaultPath)
			reader := bufio.NewReader(os.Stdin)
			vaultPath, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read vault path: %w", err)
			}
			vaultPath = strings.TrimSpace(vaultPath)

			if vaultPath == "" {
				vaultPath = defaultPath
			}

			if err := os.MkdirAll(vaultPath, 0o700); err != nil {
				return fmt.Errorf("failed to create vault directory: %w", err)
			}

			mk, err := PromptMasterPassword(true)
			if err != nil {
				return err
			}

			enc.ConfigMK(mk)

			if err := config.Save(enc, vaultPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			logger.PrintSuccess(fmt.Sprintf("Config saved. Vault path: %s\n", vaultPath))
			return nil
		},
	}

	return configCmd
}
