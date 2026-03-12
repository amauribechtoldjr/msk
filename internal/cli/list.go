package cli

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

func NewListCmd(holder *ServiceHolder) *cobra.Command {
	var (
		jsonOutput bool
		sortOrder  string
	)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "Used to list passwords from the vault.",
		RunE: func(cmd *cobra.Command, args []string) error {
			secretNames, err := holder.Service.GetSecrets()
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			if sortOrder != "" {
				slices.SortFunc(secretNames, func(a, b string) int {
					if sortOrder == "asc" {
						return cmp.Compare(a, b)
					}
					return cmp.Compare(b, a)
				})
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(secretNames)
			}

			for _, name := range secretNames {
				fmt.Println(name)
			}

			return nil
		},
	}

	listCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	listCmd.Flags().StringVarP(&sortOrder, "sort", "s", "", "Sort secrets by name (asc or desc)")

	return listCmd
}
