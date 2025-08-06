package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// BackupToS3 sauvegarde le projet vers S3
func BackupToS3(projectPath string, s3Bucket string) error {
	// Créer un fichier tar.gz du projet
	timestamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("backup-%s.tar.gz", timestamp)
	tempFile := filepath.Join(os.TempDir(), archiveName)
	defer os.Remove(tempFile)

	// Créer l'archive
	if err := execCommand("tar", "-czf", tempFile, "-C", projectPath, "."); err != nil {
		return fmt.Errorf("erreur lors de la création de l'archive: %v", err)
	}

	// Upload vers S3
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
		o.UsePathStyle = true
	})

	file, err := os.Open(tempFile)
	if err != nil {
		return fmt.Errorf("erreur ouverture fichier: %v", err)
	}
	defer file.Close()

	key := fmt.Sprintf("cli-backups/%s", archiveName)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s3Bucket,
		Key:    &key,
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("erreur upload S3: %v", err)
	}

	LogToFrontend("info", fmt.Sprintf("Backup sauvegardé vers S3: %s/%s", s3Bucket, key))
	return nil
}

// RestoreFromS3 restaure le projet depuis S3
func RestoreFromS3(s3Bucket string, projectPath string) error {
	// TODO: Implémenter la logique de restauration depuis S3
	// 1. Lister les backups disponibles
	// 2. Télécharger le backup sélectionné
	// 3. Extraire dans le projectPath
	return fmt.Errorf("restauration S3 non implémentée pour le moment")
}

// BackupToLocal sauvegarde le projet localement
func BackupToLocal(projectPath string, localPath string) error {
	// Créer le répertoire de destination s'il n'existe pas
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("erreur création répertoire: %v", err)
	}

	// Créer un fichier tar.gz du projet
	timestamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("backup-%s.tar.gz", timestamp)
	archivePath := filepath.Join(localPath, archiveName)

	// Créer l'archive directement dans le répertoire de destination
	if err := execCommand("tar", "-czf", archivePath, "-C", projectPath, "."); err != nil {
		return fmt.Errorf("erreur lors de la création de l'archive: %v", err)
	}

	LogToFrontend("info", fmt.Sprintf("Backup sauvegardé localement: %s", archivePath))
	return nil
}

// RestoreFromLocal restaure le projet depuis une sauvegarde locale
func RestoreFromLocal(localPath string, projectPath string) error {
	// Vérifier que le fichier de backup existe
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("fichier de backup introuvable: %s", localPath)
	}

	// Créer le répertoire de destination s'il n'existe pas
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("erreur création répertoire: %v", err)
	}

	// Extraire l'archive
	if err := execCommand("tar", "-xzf", localPath, "-C", projectPath); err != nil {
		return fmt.Errorf("erreur lors de l'extraction de l'archive: %v", err)
	}

	LogToFrontend("info", fmt.Sprintf("Projet restauré depuis: %s", localPath))
	return nil
}

// Helper pour copier un fichier
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}