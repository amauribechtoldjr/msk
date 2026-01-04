package main

import (
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/cli"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/storage/file"
	"github.com/amauribechtoldjr/msk/utils"
)

func main() {
	// TODO: add init config file
	store, _ := file.NewStore("./vault/")
	enc := encryption.NewArgonCrypt()
	service :=  app.NewMSKService(store, enc)

	exit := 0
	rootCmd := cli.NewMSKCmd(service)
	if err := rootCmd.Execute(); err != nil {
		utils.InfoMessage(err.Error())
		exit = 1
	}

	os.Exit(exit)
}