package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/prompt"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/spf13/cobra"
)

func NewUnlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Start the agent daemon for the current shell session",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.NewConfig()
			if err != nil {
				return err
			}

			exists, err := conf.Exists()
			if err != nil {
				return err
			}

			if !exists {
				return config.ErrConfigNotFound
			}

			mk, err := prompt.ReadMasterPassword(false)
			if err != nil {
				return err
			}

			// Copy before validation — NewVaultWithMK zeroes the input via memguard
			mkForDaemon := make([]byte, len(mk))
			copy(mkForDaemon, mk)

			tempVault := vault.NewVaultWithMK(mk)
			// mk is zeroed at this point
			if _, err := conf.Load(tempVault); err != nil {
				tempVault.DestroyMK()
				wipe.Bytes(mkForDaemon)
				return fmt.Errorf("invalid master password: %w", err)
			}
			tempVault.DestroyMK()

			exePath, err := os.Executable()
			if err != nil {
				wipe.Bytes(mkForDaemon)
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			daemonCmd := exec.Command(exePath, "daemon")
			daemonCmd.Stdout = nil
			daemonCmd.Stderr = nil
			// setDetachedProcess(daemonCmd)

			stdinPipe, err := daemonCmd.StdinPipe()
			if err != nil {
				wipe.Bytes(mkForDaemon)
				return fmt.Errorf("failed to create stdin pipe: %w", err)
			}

			if err := daemonCmd.Start(); err != nil {
				wipe.Bytes(mkForDaemon)
				return fmt.Errorf("failed to start agent: %w", err)
			}

			stdinPipe.Write(mkForDaemon)
			stdinPipe.Close()
			wipe.Bytes(mkForDaemon)

			pid := daemonCmd.Process.Pid
			sockDir := filepath.Join(os.TempDir(), fmt.Sprintf("msk-agent.%d", pid))
			sockPath := filepath.Join(sockDir, "agent.sock")

			for range 50 {
				if _, err := os.Stat(sockPath); err == nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}

			daemonCmd.Process.Release()

			fmt.Printf("MSK_AUTH_SOCK=%s", sockPath)
			return nil
		},
	}
}
