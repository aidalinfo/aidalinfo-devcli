package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Manager struct {
	Client *s3.Client
	Bucket string
}

// writeAWSCredentialsFile génère ou met à jour le fichier ~/.aws/credentials avec les clés fournies
func writeAWSCredentialsFile(accessKey, secretKey string) error {
	// Définir le chemin du fichier credentials
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du répertoire personnel : %v", err)
	}
	awsCredentialsPath := filepath.Join(homeDir, ".aws", "credentials")

	// Créer le dossier ~/.aws s'il n'existe pas
	err = os.MkdirAll(filepath.Dir(awsCredentialsPath), 0700)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du dossier .aws : %v", err)
	}

	// Lire le contenu existant du fichier credentials s'il existe
	var existingContent string
	if _, err := os.Stat(awsCredentialsPath); err == nil {
		data, err := os.ReadFile(awsCredentialsPath)
		if err != nil {
			return fmt.Errorf("erreur lors de la lecture du fichier credentials : %v", err)
		}
		existingContent = string(data)
	}

	// Vérifier si la section [aidalinfo-dev-cli] existe déjà
	sectionHeader := "[aidalinfo-devcli]"
	if existingContent != "" && containsSection(existingContent, sectionHeader) {
		log.Printf("La section %s existe déjà dans le fichier credentials. Aucune modification nécessaire.", sectionHeader)
		return nil
	}

	newSection := fmt.Sprintf(`[aidalinfo-devcli]
aws_access_key_id = %s
aws_secret_access_key = %s
`, accessKey, secretKey)

	newContent := existingContent + "\n" + newSection

	// Écrire le contenu mis à jour dans le fichier credentials
	err = os.WriteFile(awsCredentialsPath, []byte(newContent), 0600)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier credentials : %v", err)
	}

	log.Printf("La section %s a été ajoutée avec succès au fichier credentials : %s", sectionHeader, awsCredentialsPath)
	return nil
}

// containsSection vérifie si une section existe déjà dans le contenu du fichier
func containsSection(content, sectionHeader string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			return true
		}
	}
	return false
}

// NewS3Manager initialise le gestionnaire S3 en utilisant la configuration AWS par défaut
func NewS3Manager(bucket, region, endpoint string) (*S3Manager, error) {
	// Charger la configuration par défaut depuis les fichiers AWS (credentials et config)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region), // Région par défaut
		config.WithSharedConfigProfile("aidalinfo-devcli"),
	)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du chargement de la configuration AWS : %v", err)
	}
	// Initialiser le client S3 avec le point de terminaison Scaleway
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true      // Mode de chemin d'accès (obligatoire pour Scaleway)
		o.BaseEndpoint = &endpoint // Point de terminaison personnalisé
	})

	return &S3Manager{
		Client: client,
		Bucket: bucket,
	}, nil
}

// ListBackups liste les objets dans le bucket S3
func (m *S3Manager) ListBackups() ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: &m.Bucket,
	}

	result, err := m.Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la liste des objets : %v", err)
	}

	var backups []string
	for _, item := range result.Contents {
		backups = append(backups, *item.Key)
	}

	return backups, nil
}

func listTest() {
	// Appeler `getSecret` pour récupérer les informations nécessaires
	bucketName, err := getSecret("BUCKET_NAME_PROD", "Production")
	if err != nil {
		log.Fatalf("Erreur : %v", err)
	}
	accessKey, err := getSecret("SCW_ACCESS_BACKUP_ACCESS_KEY", "Production")
	if err != nil {
		log.Fatalf("Erreur : %v", err)
	}
	secretKey, err := getSecret("SCW_ACCESS_BACKUP_SECRET_KEY", "Production")
	if err != nil {
		log.Fatalf("Erreur : %v", err)
	}
	region, err := getSecret("BUCKET_REGION_PROD", "Production")
	if err != nil {
		log.Fatalf("Erreur : %v", err)
	}
	endpoint, err := getSecret("BUCKET_ENDPOINT_PROD", "Production")
	if err != nil {
		log.Fatalf("Erreur : %v", err)
	}

	// Générer le fichier credentials
	err = writeAWSCredentialsFile(accessKey, secretKey)
	if err != nil {
		log.Fatalf("Erreur lors de la génération du fichier AWS credentials : %v", err)
	}

	// Initialisation du S3Manager
	s3Manager, err := NewS3Manager(bucketName, region, endpoint)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du gestionnaire S3 : %v\n", err)
	}
	fmt.Println("S3Manager initialized")
	RunS3BucketUI(s3Manager)
}
