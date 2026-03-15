package app

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/amauribechtoldjr/msk/internal/agent"
	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
)

func BootstrapWithAuth(v vault.Vault) (Service, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	exists, err := cfg.Exists()
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("invalid config file")
	}

	var activeVault vault.Vault
	if sockPath := os.Getenv(meta.AgentSocketEnv); sockPath != "" {
		client := agent.NewClient(sockPath)
		if err := client.Ping(); err == nil {
			activeVault = client
		} else {
			os.Remove(sockPath)
			os.Remove(filepath.Dir(sockPath))
			err := v.LoadMK()
			if err != nil {
				return nil, err
			}
			activeVault = v
		}
	} else {
		err := v.LoadMK()
		if err != nil {
			return nil, err
		}
		activeVault = v
	}

	vaultPath, err := cfg.Load(activeVault)
	if err != nil {
		activeVault.DestroyMK()
		return nil, err
	}

	store, err := storage.NewStore(vaultPath)
	if err != nil {
		activeVault.DestroyMK()
		return nil, err
	}

	service := NewMSKService(store, activeVault)

	return service, nil
}
