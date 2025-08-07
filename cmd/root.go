package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	projectPath string
	branchArg   string
	Version     = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "aidalinfo-cli",
	Short: "Aidalinfo CLI - Outil de gestion des projets",
	Long:  `Aidalinfo CLI est un outil pour gérer les sous-modules Git, installer les dépendances NPM et automatiser les tâches de développement.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&projectPath, "path", ".", "Chemin du projet")
}