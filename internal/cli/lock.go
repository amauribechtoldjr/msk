package cli

import (
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/agent"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/spf13/cobra"
)

func NewLockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Stop the agent and lock the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			sockPath := os.Getenv(meta.AgentSocketEnv)
			if sockPath == "" {
				return fmt.Errorf("no agent running (MSK_AUTH_SOCK not set)")
			}

			client := agent.NewClient(sockPath)
			client.DestroyMK()

			fmt.Println("unset MSK_AUTH_SOCK")
			return nil
		},
	}
}
