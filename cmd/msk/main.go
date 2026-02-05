package main

import (
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/clip"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/storage"
)

func main() {
	if err := clip.Init(); err != nil {
		logger.RenderError(fmt.Errorf("clipboard initialization failed: %w", err))
		os.Exit(1)
	}

	store, err := storage.NewStore("./vault/")
	if err != nil {
		logger.RenderError(err)
		os.Exit(1)
	}

	enc := encryption.NewArgonCrypt()
	service := app.NewMSKService(*store, enc)

	rootCmd := cli.NewMSKCmd(*service)
	if err := rootCmd.Execute(); err != nil {
		logger.RenderError(err)
		os.Exit(1)
	}

	os.Exit(0)
}
