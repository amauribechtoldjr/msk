package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/logger"
	encryption "github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/awnumar/memguard"
)

func main() {
	memguard.CatchInterrupt()

	defer memguard.Purge()

	if err := clip.Init(); err != nil {
		logger.PrintError("clipboard initialization failed")
		os.Exit(1)
	}

	vault := encryption.NewMSKVault()

	rootCmd := cli.NewMSKCmd(vault)
	if err := rootCmd.Execute(); err != nil {
		logger.PrintError("%s\n", err)
		memguard.Purge()
		os.Exit(1)
	}

	os.Exit(0)
}
