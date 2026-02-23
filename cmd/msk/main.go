package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/awnumar/memguard"
)

func main() {
	memguard.CatchInterrupt()

	defer memguard.Purge()

	if err := clip.Init(); err != nil {
		logger.PrintError("clipboard initialization failed")
		os.Exit(1)
	}

	enc := encryption.NewArgonCrypt()

	rootCmd := cli.NewMSKCmd(enc)
	if err := rootCmd.Execute(); err != nil {
		logger.PrintError("%s\n", err)
		memguard.Purge()
		os.Exit(1)
	}

	os.Exit(0)
}
