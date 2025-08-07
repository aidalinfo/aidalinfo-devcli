package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var updateNpmCmd = &cobra.Command{
	Use:   "update-npm",
	Short: "Mettre à jour les dépendances NPM",
	Long:  `Met à jour toutes les dépendances NPM pour les submodules du projet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Mise à jour des dépendances NPM...")
		if err := backend.NpmUpdateAction(projectPath); err != nil {
			return fmt.Errorf("erreur lors de la mise à jour NPM: %w", err)
		}
		fmt.Println("Mise à jour NPM terminée avec succès!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateNpmCmd)
}