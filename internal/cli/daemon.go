package cli

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/amauribechtoldjr/msk/internal/agent"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewDaemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mk, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read master password: %w", err)
			}
			defer wipe.Bytes(mk)

			pid := os.Getpid()
			sockDir := filepath.Join(os.TempDir(), fmt.Sprintf("msk-agent.%d", pid))

			os.RemoveAll(sockDir)

			if err := os.MkdirAll(sockDir, 0o700); err != nil {
				return fmt.Errorf("failed to create socket directory: %w", err)
			}

			sockPath := filepath.Join(sockDir, "agent.sock")

			srv, err := agent.NewServer(mk, sockPath, meta.AgentTTL)
			if err != nil {
				os.RemoveAll(sockDir)
				return fmt.Errorf("failed to start agent: %w", err)
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			go func() {
				<-sigChan
				srv.Shutdown()
				os.RemoveAll(sockDir)
			}()

			srv.Serve()

			os.RemoveAll(sockDir)

			return nil
		},
	}
}
