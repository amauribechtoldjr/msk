package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/prompt"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewConfigCmd(vault vault.Vault) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure MSK vault path and master password.",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			vault.DestroyMK()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.NewConfig()
			if err != nil {
				return err
			}

			showConfig, err := cmd.Flags().GetBool("show")
			if err != nil {
				return err
			}

			if showConfig {
				logger.PrintInfo(conf.Path)
				return nil
			}

			vaultPath, err := conf.CreateVault()
			if err != nil {
				return err
			}

			mk, err := prompt.ReadMasterPassword(false)
			if err != nil {
				return err
			}
			defer wipe.Bytes(mk)

			vault.ConfigMK(mk)

			if err := conf.Save(vault, vaultPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			logger.PrintSuccess(fmt.Sprintf("Config saved. Vault path: %s\n", vaultPath))
			return nil
		},
	}

	configCmd.Flags().BoolP("show", "s", false, "Show config and session path")

	return configCmd
}
