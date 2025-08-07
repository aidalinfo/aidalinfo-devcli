package cmd

import (
	"aidalinfo-copilot/backend"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	backupType string
	s3Bucket   string
	localPath  string
	restore    bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Sauvegarder ou restaurer le projet",
	Long:  `Sauvegarde ou restaure le projet vers/depuis S3 ou un répertoire local.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if restore {
			fmt.Println("Restauration du projet...")
			
			if backupType == "s3" && s3Bucket != "" {
				if err := backend.RestoreFromS3(s3Bucket, projectPath); err != nil {
					return fmt.Errorf("erreur lors de la restauration depuis S3: %w", err)
				}
				fmt.Println("Restauration depuis S3 terminée avec succès!")
			} else if backupType == "local" && localPath != "" {
				if err := backend.RestoreFromLocal(localPath, projectPath); err != nil {
					return fmt.Errorf("erreur lors de la restauration locale: %w", err)
				}
				fmt.Println("Restauration locale terminée avec succès!")
			} else {
				return fmt.Errorf("veuillez spécifier --type (s3 ou local) et le chemin approprié")
			}
		} else {
			fmt.Println("Sauvegarde du projet...")
			
			if backupType == "s3" && s3Bucket != "" {
				if err := backend.BackupToS3(projectPath, s3Bucket); err != nil {
					return fmt.Errorf("erreur lors de la sauvegarde vers S3: %w", err)
				}
				fmt.Println("Sauvegarde vers S3 terminée avec succès!")
			} else if backupType == "local" && localPath != "" {
				if err := backend.BackupToLocal(projectPath, localPath); err != nil {
					return fmt.Errorf("erreur lors de la sauvegarde locale: %w", err)
				}
				fmt.Println("Sauvegarde locale terminée avec succès!")
			} else {
				return fmt.Errorf("veuillez spécifier --type (s3 ou local) et le chemin approprié")
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVar(&backupType, "type", "", "Type de sauvegarde (s3 ou local)")
	backupCmd.Flags().StringVar(&s3Bucket, "s3-bucket", "", "Nom du bucket S3")
	backupCmd.Flags().StringVar(&localPath, "local-path", "", "Chemin local pour la sauvegarde")
	backupCmd.Flags().BoolVar(&restore, "restore", false, "Restaurer au lieu de sauvegarder")
}