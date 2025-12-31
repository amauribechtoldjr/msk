package cli

import (
	fl "github.com/amauribechtoldjr/msk/internal/file_manager"
	u "github.com/amauribechtoldjr/msk/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Aliases: []string{"c"},
	Short: "Used to config the master key",
	Long: ``,
	Run: config,
}

func config(cmd *cobra.Command, args []string) {
	masterKey, _ := cmd.Flags().GetString("master")

	err := fl.SetUpMSK([]byte(masterKey))

	if err == nil {
		u.SuccessMessage("MSK was successfully set up on this machine.")
	} else {
		u.InfoMessage(err.Error())
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
}
