package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var npmCmd = &cobra.Command{
	Use:   "npm",
	Short: "Installer les dépendances NPM",
	Long:  `Installer toutes les dépendances NPM pour les submodules du projet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installation des dépendances NPM...")
		if err := backend.NpmAction(projectPath, true); err != nil {
			return fmt.Errorf("erreur lors de l'installation NPM: %w", err)
		}
		fmt.Println("Installation NPM terminée avec succès!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(npmCmd)
}