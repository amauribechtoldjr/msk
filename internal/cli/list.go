package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewListCmd(holder *ServiceHolder) *cobra.Command {
	listCmd := &cobra.Command{
		Use:           "list",
		Aliases:       []string{"l"},
		Short:         "Used to list passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretNames, err := holder.Service.ListSecrets()
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"ID", "Name"})

			for i, name := range secretNames {
				t.AppendRow(table.Row{i, strings.TrimSuffix(name, ".msk")})
			}

			t.Render()

			return nil
		},
	}

	return listCmd
}
