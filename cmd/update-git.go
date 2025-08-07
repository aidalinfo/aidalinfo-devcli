package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var updateGitCmd = &cobra.Command{
	Use:   "update-git",
	Short: "Mettre à jour les sous-modules (git pull)",
	Long:  `Met à jour tous les sous-modules Git en effectuant un git pull sur chacun.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Mise à jour des sous-modules Git...")
		
		submodules, err := backend.ListSubmodule(projectPath)
		if err != nil {
			return fmt.Errorf("erreur lors de la liste des submodules: %w", err)
		}
		
		if err := backend.GitUpdateAction(projectPath, submodules); err != nil {
			return fmt.Errorf("erreur lors de la mise à jour Git: %w", err)
		}
		
		fmt.Println("Mise à jour Git terminée avec succès!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateGitCmd)
}