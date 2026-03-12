package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/config"
	"github.com/amauribechtoldjr/msk/internal/session"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
)

func BootstrapWithAuth(vault vault.Vault) (Service, error) {
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

	if token := os.Getenv("MSK_SESSION"); token != "" {
		session, err := session.New()
		if err != nil {
			return nil, err
		}

		binarySession, err := session.LoadFile(token)
		if err != nil {
			fmt.Println(" 2")
			return nil, err
		}

		err = vault.LoadSession(binarySession)
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %v", err)
		}

	} else {
		err := vault.LoadMK()
		if err != nil {
			return nil, err
		}
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
