package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/storage/file"
)

func main() {
	// TODO: add init config file
	err := clip.Init()
	if err != nil {
		logger.RenderError(err)
		os.Exit(1)
	}

	store, _ := file.NewStore("./vault/")
	enc := encryption.NewArgonCrypt()
	service :=  app.NewMSKService(store, enc)

	exit := 0
	rootCmd := cli.NewMSKCmd(service)
	if err := rootCmd.Execute(); err != nil {
		logger.RenderError(err)
		exit = 1
	}

	os.Exit(exit)
}