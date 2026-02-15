package cli

import (
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type Secret struct {
	ID   int
	Name string
}

func NewListCmd(service *app.MSKService) *cobra.Command {
	listCmd := &cobra.Command{
		Use:           "list",
		Aliases:       []string{"l"},
		Short:         "Used to list passwords from the vault.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretNames, err := service.ListSecrets()
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			var users []Secret

			for i, name := range secretNames {
				users = append(users, Secret{
					ID:   i,
					Name: name,
				})
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"ID", "Name"})

			for _, u := range users {
				t.AppendRow(table.Row{u.ID, u.Name})
			}

			t.Render()

			return nil
		},
	}

	return listCmd
}
