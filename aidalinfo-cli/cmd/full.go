package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "Installation complète (submodules + npm)",
	Long:  `Effectue une installation complète en installant les submodules Git et les dépendances NPM.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installation complète en cours...")
		
		fmt.Println("1. Installation des submodules...")
		if err := backend.SubmoduleAction(projectPath); err != nil {
			return fmt.Errorf("erreur lors de l'installation des submodules: %w", err)
		}
		
		fmt.Println("2. Installation des dépendances NPM...")
		if err := backend.NpmAction(projectPath, true); err != nil {
			return fmt.Errorf("erreur lors de l'installation NPM: %w", err)
		}
		
		fmt.Println("Installation complète terminée avec succès!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fullCmd)
}