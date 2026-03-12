package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/vault"
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
				exists, err := conf.Exists()
				if err != nil {
					return err
				}

				if !exists {
					return config.ErrConfigNotFound
				}

				logger.PrintInfo(conf.Path)
				logger.Lb()
				return nil
			}

			exists, err := conf.Exists()
			if err != nil {
				return err
			}

			var shouldOverwrite bool
			if exists {
				shouldOverwrite, err = conf.CheckOverwrite()
				if err != nil {
					return err
				}

				if !shouldOverwrite {
					logger.PrintSuccess("Config unchanged\n")
					return nil
				}
			}

			vaultPath, err := conf.CreateVault()
			if err != nil {
				return err
			}

			err = vault.LoadMK()
			if err != nil {
				return err
			}

			if err := conf.Save(vault, vaultPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			logger.PrintSuccess(fmt.Sprintf("Vault path created successfully at: %s\n", vaultPath))
			return nil
		},
	}

	configCmd.Flags().BoolP("show", "s", false, "Show config and session path")

	return configCmd
}
