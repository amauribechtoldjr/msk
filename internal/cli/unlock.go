package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/prompt"
	"github.com/amauribechtoldjr/msk/internal/session"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewUnlockCmd(vault vault.Vault) *cobra.Command {
	return &cobra.Command{
		Use:           "unlock",
		Short:         "Unlock the vault for the current shell session",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mk, err := prompt.PromptMasterPassword(false)
			if err != nil {
				return err
			}

			vault.ConfigMK(mk)
			defer vault.DestroyMK()

			conf, err := config.NewConfig()
			if err != nil {
				return err
			}

			if _, err := conf.Load(vault); err != nil {
				return fmt.Errorf("invalid master password: %w", err)
			}

			s, err := session.New()
			if err != nil {
				return fmt.Errorf("failed to initialize session: %w", err)
			}

			token, err := s.GetSessionToken()
			defer wipe.Bytes(token)

			encodedToken := hex.EncodeToString(token)
			sealedSession, err := vault.CreateSession(token)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			err = s.StoreSession(sealedSession)
			if err != nil {
				return fmt.Errorf("failed to store session: %w", err)
			}

			fmt.Printf("export MSK_SESSION=%s\n", encodedToken)
			return nil
		},
	}
}
