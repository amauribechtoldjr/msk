package cli

import (
	"github.com/amauribechtoldjr/msk/internal/build"
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print the version information.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.PrintInfo(build.Version)
		},
	}
}
