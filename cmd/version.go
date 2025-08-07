package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Afficher la version",
	Long:  `Affiche la version actuelle d'aidalinfo-cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aidalinfo-cli version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}