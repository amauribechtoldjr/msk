package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
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

	rootCmd := cli.NewMSKCmd()
	if err := rootCmd.Execute(); err != nil {
		memguard.Purge()
		os.Exit(1)
	}

	os.Exit(0)
}
