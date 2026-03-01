package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/config"
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
			mk, err := PromptMasterPassword(false)
			if err != nil {
				return err
			}

			mkCopy := make([]byte, len(mk))
			copy(mkCopy, mk)
			defer wipe.Bytes(mkCopy)

			vault.ConfigMK(mk)
			defer vault.DestroyMK()

			if _, err := config.Load(vault); err != nil {
				return fmt.Errorf("invalid master password: %w", err)
			}

			sess, err := session.New()
			if err != nil {
				return fmt.Errorf("failed to initialize session: %w", err)
			}

			token, err := sess.Create(mkCopy)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			fmt.Printf("export MSK_SESSION=%s\n", token)
			return nil
		},
	}
}
