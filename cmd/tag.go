package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	tagName    string
	tagMessage string
	submodule  string
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Créer un tag Git",
	Long:  `Créer un nouveau tag Git pour un submodule spécifique ou le projet principal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tagName == "" {
			return fmt.Errorf("le nom du tag est requis (--name)")
		}

		targetPath := projectPath
		if submodule != "" {
			targetPath = fmt.Sprintf("%s/%s", projectPath, submodule)
		}

		fmt.Printf("Création du tag '%s' dans %s...\n", tagName, targetPath)
		
		if err := backend.CreateTag(targetPath, tagName, tagMessage); err != nil {
			return fmt.Errorf("erreur lors de la création du tag: %w", err)
		}
		
		fmt.Printf("Tag '%s' créé avec succès!\n", tagName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().StringVar(&tagName, "name", "", "Nom du tag à créer")
	tagCmd.Flags().StringVar(&tagMessage, "message", "", "Message du tag")
	tagCmd.Flags().StringVar(&submodule, "submodule", "", "Submodule spécifique (sinon utilise le projet principal)")
}