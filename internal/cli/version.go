package cli

import (
	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print the version information.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.PrintInfo(meta.Version)
		},
	}
}
