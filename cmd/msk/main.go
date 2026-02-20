package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/awnumar/memguard"
)

func main() {
	memguard.CatchInterrupt()

	defer memguard.Purge()

	if err := clip.Init(); err != nil {
		logger.PrintError("clipboard initialization failed")
		os.Exit(1)
	}

	store, err := storage.NewStore("./vault/")
	if err != nil {
		logger.PrintError("vault initialization failed")
		os.Exit(1)
	}

	enc := encryption.NewArgonCrypt()
	service := app.NewMSKService(store, enc)

	rootCmd := cli.NewMSKCmd(service)
	if err := rootCmd.Execute(); err != nil {
		logger.PrintError("%s\n", err)
		memguard.Purge()
		os.Exit(1)
	}

	os.Exit(0)
}
