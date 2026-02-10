package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ListMySQLDatabases liste les bases de données disponibles sur un serveur MySQL
func ListMySQLDatabases(ctx context.Context, mysqlHost, mysqlPort, mysqlUser, mysqlPassword string) ([]string, error) {
	// Construire la commande mysql pour lister les bases
	args := []string{
		"-h", mysqlHost,
		"-P", mysqlPort,
		"-u", mysqlUser,
	}

	if mysqlPassword != "" {
		args = append(args, fmt.Sprintf("-p%s", mysqlPassword))
	}

	// Exécuter SHOW DATABASES
	args = append(args, "-e", "SHOW DATABASES;", "--skip-column-names", "--batch")

	cmd := exec.Command("mysql", args...)
	output, err := cmd.Output()
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur listing databases MySQL: %v", err))
		return nil, fmt.Errorf("erreur listing databases MySQL: %v", err)
	}

	// Parser la sortie pour obtenir la liste des bases
	lines := strings.Split(string(output), "\n")
	var databases []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Exclure les bases système MySQL
		if line != "" && line != "information_schema" && line != "performance_schema" && line != "mysql" && line != "sys" {
			databases = append(databases, line)
		}
	}

	return databases, nil
}

// DumpMySQLDatabase crée un dump d'une base MySQL
func DumpMySQLDatabase(ctx context.Context, mysqlHost, mysqlPort, mysqlUser, mysqlPassword, database string) (string, error) {
	tmpDir, err := getUserTmpDir()
	if err != nil {
		return "", err
	}

	// Créer un fichier temporaire pour le dump
	tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("mysql-dump-%s-*.sql.gz", database))
	if err != nil {
		return "", fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Prépare la commande mysqldump avec compression gzip
	args := []string{
		"-h", mysqlHost,
		"-P", mysqlPort,
		"-u", mysqlUser,
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		database,
	}

	if mysqlPassword != "" {
		// Réorganiser pour mettre le password avant les autres options
		args = []string{
			"-h", mysqlHost,
			"-P", mysqlPort,
			"-u", mysqlUser,
			fmt.Sprintf("-p%s", mysqlPassword),
			"--single-transaction",
			"--routines",
			"--triggers",
			"--events",
			database,
		}
	}

	LogToFrontend("info", fmt.Sprintf("Création du dump de la base MySQL %s...", database))

	// Utiliser mysqldump avec pipe vers gzip
	cmdDump := exec.Command("mysqldump", args...)
	cmdGzip := exec.Command("gzip", "-c")

	// Connecter la sortie de mysqldump à l'entrée de gzip
	pipe, err := cmdDump.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("erreur création pipe: %v", err)
	}
	cmdGzip.Stdin = pipe

	// Rediriger la sortie de gzip vers le fichier
	outFile, err := os.Create(tmpFilePath)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier sortie: %v", err)
	}
	defer outFile.Close()
	cmdGzip.Stdout = outFile

	// Démarrer les commandes
	if err := cmdGzip.Start(); err != nil {
		return "", fmt.Errorf("erreur démarrage gzip: %v", err)
	}
	if err := cmdDump.Start(); err != nil {
		return "", fmt.Errorf("erreur démarrage mysqldump: %v", err)
	}

	// Attendre la fin des commandes
	if err := cmdDump.Wait(); err != nil {
		os.Remove(tmpFilePath)
		LogToFrontend("error", fmt.Sprintf("Erreur mysqldump: %v", err))
		return "", fmt.Errorf("erreur mysqldump: %v", err)
	}
	pipe.Close()

	if err := cmdGzip.Wait(); err != nil {
		os.Remove(tmpFilePath)
		LogToFrontend("error", fmt.Sprintf("Erreur gzip: %v", err))
		return "", fmt.Errorf("erreur gzip: %v", err)
	}

	LogToFrontend("success", fmt.Sprintf("Dump MySQL de %s créé avec succès", database))
	return tmpFilePath, nil
}

// TransferMySQLDatabase transfère une base de données entre deux serveurs MySQL
func TransferMySQLDatabase(ctx context.Context, sourceHost, sourcePort, sourceUser, sourcePassword,
	destHost, destPort, destUser, destPassword, database string, dropExisting bool) error {

	LogToFrontend("info", fmt.Sprintf("Début du transfert de la base MySQL %s", database))

	// Étape 1: Créer le dump de la source
	dumpFile, err := DumpMySQLDatabase(ctx, sourceHost, sourcePort, sourceUser, sourcePassword, database)
	if err != nil {
		return fmt.Errorf("erreur création dump: %v", err)
	}
	defer os.Remove(dumpFile)

	// Étape 2: Si dropExisting, supprimer la base de destination si elle existe
	if dropExisting {
		LogToFrontend("info", fmt.Sprintf("Suppression de la base %s sur le serveur de destination...", database))
		dropArgs := []string{
			"-h", destHost,
			"-P", destPort,
			"-u", destUser,
		}
		if destPassword != "" {
			dropArgs = append(dropArgs, fmt.Sprintf("-p%s", destPassword))
		}
		dropArgs = append(dropArgs, "-e", fmt.Sprintf("DROP DATABASE IF EXISTS `%s`; CREATE DATABASE `%s`;", database, database))

		cmdDrop := exec.Command("mysql", dropArgs...)
		if err := cmdDrop.Run(); err != nil {
			LogToFrontend("warn", fmt.Sprintf("Impossible de recréer la base: %v", err))
			// On continue quand même, la base existe peut-être déjà
		}
	}

	// Étape 3: Restaurer sur la destination
	LogToFrontend("info", fmt.Sprintf("Restauration de %s sur le serveur MySQL de destination...", database))

	// Décompresser et restaurer
	cmdGunzip := exec.Command("gunzip", "-c", dumpFile)
	cmdMysql := exec.Command("mysql",
		"-h", destHost,
		"-P", destPort,
		"-u", destUser,
		fmt.Sprintf("-p%s", destPassword),
		database,
	)

	// Si pas de password, ajuster les arguments
	if destPassword == "" {
		cmdMysql = exec.Command("mysql",
			"-h", destHost,
			"-P", destPort,
			"-u", destUser,
			database,
		)
	}

	// Connecter gunzip à mysql via pipe
	pipe, err := cmdGunzip.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erreur création pipe: %v", err)
	}
	cmdMysql.Stdin = pipe
	cmdMysql.Stdout = os.Stdout
	cmdMysql.Stderr = os.Stderr

	// Démarrer les commandes
	if err := cmdMysql.Start(); err != nil {
		return fmt.Errorf("erreur démarrage mysql: %v", err)
	}
	if err := cmdGunzip.Start(); err != nil {
		return fmt.Errorf("erreur démarrage gunzip: %v", err)
	}

	// Attendre la fin
	if err := cmdGunzip.Wait(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur gunzip: %v", err))
		return fmt.Errorf("erreur gunzip: %v", err)
	}
	pipe.Close()

	if err := cmdMysql.Wait(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur mysql restore: %v", err))
		return fmt.Errorf("erreur mysql restore: %v", err)
	}

	LogToFrontend("success", fmt.Sprintf("Transfert MySQL de %s terminé avec succès", database))
	return nil
}

// RestoreMySQLBackup télécharge un backup S3 et le restaure dans MySQL
func RestoreMySQLBackup(ctx context.Context, creds S3Credentials, s3Path string, mysqlHost, mysqlPort, mysqlUser, mysqlPassword, database string) error {
	bucket, region, endpoint := resolveS3Config(creds)
	objectName := s3Path
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKey, creds.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
		o.UsePathStyle = true
	})
	presignedURL, err := generatePresignedURL(ctx, client, bucket, objectName)
	if err != nil {
		return err
	}

	// Récupère la taille du fichier pour la progression
	head, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &bucket, Key: &objectName})
	var totalSize int64 = 0
	if err == nil && head.ContentLength != nil {
		totalSize = *head.ContentLength
		LogToFrontend("info", fmt.Sprintf("Taille du backup MySQL à télécharger: %.2f MB", float64(totalSize)/(1024*1024)))
	}

	respBody, err := downloadWithRetry(presignedURL, 3, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("erreur téléchargement HTTP: %v", err)
	}
	defer respBody.Close()

	tmpDir, err := getUserTmpDir()
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(tmpDir, "mysql-backup-*.sql.gz")
	if err != nil {
		return fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Progression du téléchargement
	progressReader := &progressReaderWithLog{
		r:     respBody,
		total: totalSize,
	}
	LogToFrontend("info", "Début du téléchargement du backup MySQL...")
	_, err = io.Copy(tmpFile, progressReader)
	if err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}
	LogToFrontend("success", "Téléchargement du backup MySQL terminé.")

	// Créer la base de données si elle n'existe pas
	LogToFrontend("info", fmt.Sprintf("Création de la base de données %s si nécessaire...", database))
	createDbArgs := []string{
		"-h", mysqlHost,
		"-P", mysqlPort,
		"-u", mysqlUser,
	}
	if mysqlPassword != "" {
		createDbArgs = append(createDbArgs, fmt.Sprintf("-p%s", mysqlPassword))
	}
	createDbArgs = append(createDbArgs, "-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", database))

	cmdCreateDb := exec.Command("mysql", createDbArgs...)
	if err := cmdCreateDb.Run(); err != nil {
		LogToFrontend("warn", fmt.Sprintf("Impossible de créer la base: %v", err))
	}

	// Restaurer le backup
	LogToFrontend("info", "Début de la restauration MySQL...")

	// Décompresser et restaurer
	cmdGunzip := exec.Command("gunzip", "-c", tmpFile.Name())
	cmdMysql := exec.Command("mysql",
		"-h", mysqlHost,
		"-P", mysqlPort,
		"-u", mysqlUser,
		database,
	)

	if mysqlPassword != "" {
		cmdMysql = exec.Command("mysql",
			"-h", mysqlHost,
			"-P", mysqlPort,
			"-u", mysqlUser,
			fmt.Sprintf("-p%s", mysqlPassword),
			database,
		)
	}

	// Connecter gunzip à mysql via pipe
	pipe, err := cmdGunzip.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erreur création pipe: %v", err)
	}
	cmdMysql.Stdin = pipe
	cmdMysql.Stdout = os.Stdout
	cmdMysql.Stderr = os.Stderr

	// Démarrer les commandes
	if err := cmdMysql.Start(); err != nil {
		return fmt.Errorf("erreur démarrage mysql: %v", err)
	}
	if err := cmdGunzip.Start(); err != nil {
		return fmt.Errorf("erreur démarrage gunzip: %v", err)
	}

	// Attendre la fin
	if err := cmdGunzip.Wait(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur gunzip: %v", err))
		return fmt.Errorf("erreur gunzip: %v", err)
	}
	pipe.Close()

	if err := cmdMysql.Wait(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur mysql restore: %v", err))
		return fmt.Errorf("erreur mysql restore: %v", err)
	}

	LogToFrontend("success", "Restauration MySQL terminée avec succès.")
	return nil
}

// TestMySQLConnection teste la connexion à un serveur MySQL
func TestMySQLConnection(ctx context.Context, mysqlHost, mysqlPort, mysqlUser, mysqlPassword string) error {
	args := []string{
		"-h", mysqlHost,
		"-P", mysqlPort,
		"-u", mysqlUser,
	}

	if mysqlPassword != "" {
		args = append(args, fmt.Sprintf("-p%s", mysqlPassword))
	}

	// Tester avec une simple requête SELECT 1
	args = append(args, "-e", "SELECT 1;", "--batch")

	cmd := exec.Command("mysql", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("échec de la connexion MySQL: %v", err)
	}

	return nil
}
