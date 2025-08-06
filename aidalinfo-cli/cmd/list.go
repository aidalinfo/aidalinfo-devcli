package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lister les submodules",
	Long:  `Affiche la liste de tous les submodules Git du projet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		submodules, err := backend.ListSubmodule(projectPath)
		if err != nil {
			return fmt.Errorf("erreur lors de la liste des submodules: %w", err)
		}

		if len(submodules) == 0 {
			fmt.Println("Aucun submodule trouvé dans ce projet.")
			return nil
		}

		fmt.Println("Submodules du projet:")
		fmt.Println("---------------------")
		for i, submodule := range submodules {
			cleanName, err := backend.CleanSubmoduleName(submodule)
			if err != nil {
				cleanName = submodule
			}
			
			currentBranch, err := backend.GetCurrentBranch(fmt.Sprintf("%s/%s", projectPath, submodule))
			if err != nil {
				currentBranch = "unknown"
			}
			
			fmt.Printf("%d. %s (branche: %s)\n", i+1, cleanName, currentBranch)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}