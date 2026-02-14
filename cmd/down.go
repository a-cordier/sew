package cmd

import (
	"fmt"

	"github.com/a-cordier/sew/internal/kind"
	sewlog "github.com/a-cordier/sew/internal/log"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Delete the cluster defined in the config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return sewlog.WithSpinner(
			fmt.Sprintf("Deleting cluster %q", cfg.Kind.Name),
			func() error {
				return kind.Delete(cfg.Kind.Name)
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
