package cli

import (
	"fmt"

	"github.com/amauribechtoldjr/msk/internal/session"
	"github.com/spf13/cobra"
)

func NewLockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock the vault and end the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			sess, err := session.New()
			if err != nil {
				return fmt.Errorf("failed to initialize session: %w", err)
			}

			if err := sess.Destroy(); err != nil {
				return fmt.Errorf("failed to destroy session: %w", err)
			}

			fmt.Println("unset MSK_SESSION")
			return nil
		},
	}
}
