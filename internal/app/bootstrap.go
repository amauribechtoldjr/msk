package app

import (
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/prompt"
	"github.com/amauribechtoldjr/msk/internal/session"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

func BootstrapWithAuth(vault vault.Vault) (Service, error) {
	if token := os.Getenv("MSK_SESSION"); token != "" {
		session, err := session.New()
		if err != nil {
			return nil, err
		}

		binarySession, err := session.LoadFile(token)
		if err != nil {
			return nil, err
		}

		if err = vault.LoadSession(binarySession); err != nil {
			return nil, fmt.Errorf("failed to load session: %v", err)
		}
	} else {
		mk, err := prompt.PromptMasterPassword(false)
		if err != nil {
			return nil, err
		}
		defer wipe.Bytes(mk)
		vault.ConfigMK(mk)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	vaultPath, err := cfg.Load(vault)
	if err != nil {
		vault.DestroyMK()
		return nil, err
	}

	store, err := storage.NewStore(vaultPath)
	if err != nil {
		vault.DestroyMK()
		return nil, err
	}

	service := NewMSKService(store, vault)

	return service, nil
}
