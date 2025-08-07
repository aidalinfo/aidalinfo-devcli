package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var npmFlag bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installer les submodules",
	Long:  `Installer les submodules Git avec possibilité de spécifier des branches.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		branches := []string{}
		if branchArg != "" {
			branches = strings.Fields(branchArg)
			fmt.Printf("Installation des sous-modules avec les branches : %v\n", branches)
		} else {
			fmt.Println("Installation des sous-modules avec les branches par défaut")
		}

		if err := backend.SubmoduleAction(projectPath, branches...); err != nil {
			return fmt.Errorf("erreur lors de l'installation des submodules: %w", err)
		}

		if npmFlag {
			fmt.Println("Installation des dépendances NPM...")
			if err := backend.NpmAction(projectPath, true); err != nil {
				return fmt.Errorf("erreur lors de l'installation NPM: %w", err)
			}
		}

		fmt.Println("Installation terminée avec succès!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&branchArg, "branch", "", "Spécifier la ou les branches (séparées par un espace)")
	installCmd.Flags().BoolVar(&npmFlag, "npm", false, "Installer aussi les dépendances npm")
}