package backend

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const S3BaseURL = "s3.fr-par.scw.cloud"
const S3Region = "fr-par"

// S3Credentials structure for passing credentials from frontend
type S3Credentials struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// BackupInfo structure for frontend (nom, taille, date)
type BackupInfo struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified string `json:"lastModified"`
}

// ListBackupsWithCreds liste les backups S3 avec infos (nom, taille, date) - exclut les backups Glacier
func ListBackupsWithCreds(ctx context.Context, creds S3Credentials, s3Dir string) ([]BackupInfo, error) {
	bucket := "backup-global"
	prefix := s3Dir
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKey, creds.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
		o.UsePathStyle = true
	})
	var files []BackupInfo

	// Utilise ListObjectsV2 avec tri par date de modification (plus récent en premier)
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
		// Note: AWS S3 ne supporte pas le filtrage par StorageClass dans ListObjectsV2
		// mais on peut optimiser en récupérant les métadonnées seulement pour les objets standards
	}

	paginator := s3.NewListObjectsV2Paginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du listing S3: %v", err)
		}
		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, "/") {
				// Vérification rapide de la classe de stockage
				storageClass := string(obj.StorageClass)

				// Filtre directement sans logging excessif
				isStandard := storageClass == "" || storageClass == "STANDARD"

				if isStandard {
					files = append(files, BackupInfo{
						Name:         lastPathPart(*obj.Key),
						Size:         derefInt64(obj.Size),
						LastModified: obj.LastModified.Format("2006-01-02 15:04:05"),
					})
				}
			}
		}
	}

	// Trie les backups du plus récent au plus ancien
	sort.Slice(files, func(i, j int) bool {
		// Parse les dates pour pouvoir les comparer
		timeI, errI := time.Parse("2006-01-02 15:04:05", files[i].LastModified)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05", files[j].LastModified)

		// Si une des dates ne peut pas être parsée, utilise l'ordre alphabétique inverse
		if errI != nil || errJ != nil {
			return files[i].LastModified > files[j].LastModified
		}

		// Retourne true si timeI est plus récent que timeJ (ordre décroissant)
		return timeI.After(timeJ)
	})

	return files, nil
}

// ListBackupsWithCredsPaged liste les backups S3 paginés (10 par page)
func ListBackupsWithCredsPaged(ctx context.Context, creds S3Credentials, s3Dir string, page int, pageSize int) ([]BackupInfo, int, error) {
	all, err := ListBackupsWithCreds(ctx, creds, s3Dir)
	if err != nil {
		return nil, 0, err
	}
	total := len(all)
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start > total {
		return []BackupInfo{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func lastPathPart(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// generatePresignedURL génère une URL de téléchargement temporaire S3 (20 minutes)
func generatePresignedURL(ctx context.Context, client *s3.Client, bucket, key string) (string, error) {
	presigner := s3.NewPresignClient(client)
	presignInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	presigned, err := presigner.PresignGetObject(ctx, presignInput, func(opts *s3.PresignOptions) {
		// Utiliser une durée de validité de 20 minutes comme demandé
		opts.Expires = 20 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("erreur génération presigned URL: %v", err)
	}
	return presigned.URL, nil
}

// Télécharge une URL HTTP avec retry et timeout long
func downloadWithRetry(url string, maxAttempts int, timeout time.Duration) (io.ReadCloser, error) {
	var lastErr error

	// Utilise un client HTTP avec des timeouts plus longs et des paramètres optimisés pour les gros fichiers
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			// Augmente les timeouts pour les gros fichiers
			ResponseHeaderTimeout: 30 * time.Second,
			TLSHandshakeTimeout:   20 * time.Second,
			// Permet plus de temps pour établir une connexion
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		LogToFrontend("debug", fmt.Sprintf("Tentative de téléchargement #%d/%d", attempt, maxAttempts))

		// Crée une requête pour pouvoir personnaliser les headers
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// Ajoute des headers pour optimiser le téléchargement
		req.Header.Set("Connection", "keep-alive")

		// Exécute la requête
		resp, err := client.Do(req)

		if err == nil && resp.StatusCode == 200 {
			LogToFrontend("debug", "Connexion établie avec succès, début du téléchargement")
			return resp.Body, nil
		}

		// Gestion des erreurs
		if resp != nil {
			statusText := resp.Status
			resp.Body.Close()
			LogToFrontend("warn", fmt.Sprintf("Échec tentative #%d: HTTP status %s", attempt, statusText))
			lastErr = fmt.Errorf("HTTP status: %s", statusText)
		} else if err != nil {
			LogToFrontend("warn", fmt.Sprintf("Échec tentative #%d: %v", attempt, err))
			lastErr = err
		}

		// Pause exponentielle entre les tentatives
		backoffTime := time.Duration(attempt*attempt) * 2 * time.Second
		LogToFrontend("debug", fmt.Sprintf("Attente de %v avant la prochaine tentative", backoffTime))
		time.Sleep(backoffTime)
	}

	return nil, fmt.Errorf("échec téléchargement après %d tentatives: %v", maxAttempts, lastErr)
}

// DownloadBackupWithCreds télécharge un backup S3 avec credentials fournis (bucket privé, signature S3 via AWS SDK)
func DownloadBackupWithCreds(ctx context.Context, creds S3Credentials, s3Path, destPath string) error {
	bucket := "backup-global"
	objectName := s3Path
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKey, creds.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
		o.UsePathStyle = true
	})
	getObjInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objectName,
	}
	resp, err := client.GetObject(ctx, getObjInput)
	if err != nil {
		return fmt.Errorf("erreur téléchargement S3: %v", err)
	}
	defer resp.Body.Close()
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %v", err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}
	return nil
}

// getUserTmpDir retourne un dossier temporaire sécurisé compatible avec tous les OS
func getUserTmpDir() (string, error) {
	// Utilise le dossier temporaire du système (compatible Windows, macOS, Linux)
	baseTmpDir := os.TempDir()
	
	// Crée un sous-dossier spécifique à l'application pour éviter les conflits
	appTmpDir := fmt.Sprintf("%s/aidalinfo-cli-tmp", baseTmpDir)
	
	// Vérifie si le dossier existe, sinon le crée
	if _, err := os.Stat(appTmpDir); os.IsNotExist(err) {
		err = os.MkdirAll(appTmpDir, 0o755)
		if err != nil {
			return "", fmt.Errorf("erreur création dossier tmp: %v", err)
		}
	}
	
	return appTmpDir, nil
}

// RestoreMongoBackup télécharge un backup S3 et le restaure dans MongoDB
func RestoreMongoBackup(ctx context.Context, creds S3Credentials, s3Path string, mongoHost, mongoPort, mongoUser, mongoPassword string) error {
	bucket := "backup-global"
	objectName := s3Path
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKey, creds.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
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
		LogToFrontend("info", fmt.Sprintf("Taille du backup à télécharger: %.2f MB", float64(totalSize)/(1024*1024)))
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
	tmpFile, err := os.CreateTemp(tmpDir, "mongo-backup-*.bson.gz")
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
	LogToFrontend("info", "Début du téléchargement du backup MongoDB...")
	_, err = io.Copy(tmpFile, progressReader)
	if err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}
	LogToFrontend("success", "Téléchargement du backup MongoDB terminé.")

	LogToFrontend("debug", fmt.Sprintf("mongoHost=%s, mongoPort=%s, mongoUser=%s, mongoPassword=%s", mongoHost, mongoPort, mongoUser, mongoPassword))

	// Prépare la commande mongorestore
	args := []string{"--gzip", "--archive=" + tmpFile.Name(), "--host", mongoHost, "--port", mongoPort}
	if mongoUser != "" {
		args = append(args, "--username", mongoUser)
		// Ajouter la base d'authentification (par défaut admin)
		args = append(args, "--authenticationDatabase", "admin")
	}
	if mongoPassword != "" {
		args = append(args, "--password", mongoPassword)
	}
	LogToFrontend("info", "Début de la restauration mongorestore...")
	cmd := exec.Command("mongorestore", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	LogToFrontend("debug", fmt.Sprintf("mongorestore args: %v", args))
	if err := cmd.Run(); err != nil {
		LogToFrontend("error", fmt.Sprintf("mongorestore error: %v", err))
		return fmt.Errorf("erreur restauration mongorestore: %v", err)
	}
	LogToFrontend("success", "Restauration mongorestore terminée avec succès.")
	return nil
}

// RestoreS3Backup télécharge un backup S3 (tar.gz) et le restaure dans un S3 local (MinIO ou autre)
func RestoreS3Backup(ctx context.Context, cloudCreds S3Credentials, localCreds S3Credentials, s3Path, s3Host, s3Port, s3Region string, s3UseHttps bool) error {
	bucket := "backup-global"
	objectName := s3Path

	LogToFrontend("debug", "RestoreS3Backup: Début de la restauration S3")
	LogToFrontend("debug", fmt.Sprintf("Paramètres: bucket=%s, objectName=%s, s3Host=%s, s3Port=%s", bucket, objectName, s3Host, s3Port))

	// Utilise les credentials cloud pour télécharger le backup
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cloudCreds.AccessKey, cloudCreds.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
		o.UsePathStyle = true
	})

	// Récupère la taille du fichier avant de commencer
	var totalSize int64 = 0
	if head, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &bucket, Key: &objectName}); err == nil && head.ContentLength != nil {
		totalSize = *head.ContentLength
		LogToFrontend("info", fmt.Sprintf("Taille du fichier à télécharger: %.2f MB", float64(totalSize)/(1024*1024)))
	} else {
		LogToFrontend("warn", fmt.Sprintf("Impossible de récupérer la taille du fichier: %v", err))
	}

	LogToFrontend("debug", "Génération de l'URL présignée...")
	presignedURL, err := generatePresignedURL(ctx, client, bucket, objectName)
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur génération URL présignée: %v", err))
		return err
	}
	LogToFrontend("debug", "URL présignée générée avec succès (valide 12 heures)")

	// Prépare le fichier temporaire avant de commencer le téléchargement
	tmpDir, err := getUserTmpDir()
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur récupération dossier tmp: %v", err))
		return err
	}
	LogToFrontend("debug", fmt.Sprintf("Dossier temporaire: %s", tmpDir))

	tmpFile, err := os.CreateTemp(tmpDir, "s3-backup-*.tar.gz")
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur création fichier temporaire: %v", err))
		return fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	LogToFrontend("debug", fmt.Sprintf("Fichier temporaire créé: %s", tmpFilePath))

	// Ferme le fichier pour le rouvrir en mode append plus tard
	tmpFile.Close()

	// Vérification de l'espace disque disponible
	df, err := exec.Command("df", "-h", tmpDir).Output()
	if err == nil {
		LogToFrontend("debug", fmt.Sprintf("Espace disque disponible: %s", string(df)))
	}

	LogToFrontend("info", "Début du téléchargement, cela peut prendre plusieurs minutes...")

	// Utilise un context avec timeout plus long pour les gros fichiers
	copyCtx, cancel := context.WithTimeout(ctx, 4*time.Hour)
	defer cancel()

	// Canal pour récupérer le résultat du téléchargement
	type downloadResult struct {
		written int64
		err     error
	}
	resultChan := make(chan downloadResult, 1)

	// Lance le téléchargement dans une goroutine avec gestion avancée
	go func() {
		var totalWritten int64 = 0
		var downloadErr error

		// Nombre maximum de tentatives pour le téléchargement complet
		maxAttempts := 5
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if attempt > 1 {
				LogToFrontend("warn", fmt.Sprintf("Tentative #%d de reprise du téléchargement...", attempt))
				// Regenere un nouveau lien présigné pour chaque nouvelle tentative
				presignedURL, err = generatePresignedURL(ctx, client, bucket, objectName)
				if err != nil {
					downloadErr = fmt.Errorf("erreur regénération URL présignée: %v", err)
					break
				}
			}

			// Ouvre le fichier en mode append pour reprendre le téléchargement
			tmpFile, err = os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				downloadErr = fmt.Errorf("erreur ouverture fichier temporaire: %v", err)
				break
			}

			// Si nous avons déjà téléchargé une partie, utilisons un Range request
			var respBody io.ReadCloser
			if totalWritten > 0 {
				// Crée une requête avec Range header
				req, err := http.NewRequest("GET", presignedURL, nil)
				if err != nil {
					tmpFile.Close()
					downloadErr = fmt.Errorf("erreur création requête: %v", err)
					break
				}

				// Spécifie à partir d'où reprendre le téléchargement
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", totalWritten))
				LogToFrontend("info", fmt.Sprintf("Reprise du téléchargement à partir de %.2f MB", float64(totalWritten)/(1024*1024)))

				// Utilise un client HTTP avec timeout long
				client := &http.Client{Timeout: 2 * time.Hour}
				resp, err := client.Do(req)
				if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 206) {
					statusText := "erreur"
					if resp != nil {
						statusText = resp.Status
						resp.Body.Close()
					}
					tmpFile.Close()
					downloadErr = fmt.Errorf("erreur reprise téléchargement: %v, status: %s", err, statusText)
					// Attente avant nouvelle tentative
					time.Sleep(5 * time.Second)
					continue
				}
				respBody = resp.Body
			} else {
				// Premier téléchargement
				respBody, err = downloadWithRetry(presignedURL, 3, 2*time.Hour)
				if err != nil {
					tmpFile.Close()
					downloadErr = fmt.Errorf("erreur téléchargement HTTP: %v", err)
					// Attente avant nouvelle tentative
					time.Sleep(5 * time.Second)
					continue
				}
			}

			// Configure le lecteur avec suivi de progression
			progressReader := &progressReaderWithLog{
				r:          respBody,
				total:      totalSize,
				read:       0,
				last:       0,
				lastUpdate: time.Time{},
				lastLog:    time.Time{},
			}

			// Copie les données
			written, err := io.Copy(tmpFile, progressReader)
			respBody.Close()
			tmpFile.Close()

			if err != nil {
				LogToFrontend("warn", fmt.Sprintf("Erreur pendant le téléchargement: %v (écrit %.2f MB)",
					err, float64(totalWritten+written)/(1024*1024)))
				// Enregistre ce qui a été écrit jusqu'à présent
				totalWritten += written
				if attempt == maxAttempts {
					downloadErr = fmt.Errorf("erreur téléchargement après %d tentatives: %v", maxAttempts, err)
				}
				// Attente avant nouvelle tentative
				time.Sleep(5 * time.Second)
				continue
			}

			// Téléchargement réussi
			totalWritten += written
			LogToFrontend("success", fmt.Sprintf("Téléchargement terminé avec succès! Écrit: %.2f MB", float64(totalWritten)/(1024*1024)))
			break
		}

		resultChan <- downloadResult{written: totalWritten, err: downloadErr}
	}()

	// Attend le résultat ou le timeout
	var written int64
	var downloadErr error
	select {
	case result := <-resultChan:
		written = result.written
		downloadErr = result.err
		LogToFrontend("debug", "Téléchargement terminé")
	case <-copyCtx.Done():
		LogToFrontend("error", "TIMEOUT lors du téléchargement après 4 heures")
		return fmt.Errorf("timeout lors du téléchargement après 4 heures")
	}

	if downloadErr != nil {
		LogToFrontend("error", fmt.Sprintf("ERREUR téléchargement: %v (écrit: %.2f MB)", downloadErr, float64(written)/(1024*1024)))
		return fmt.Errorf("erreur téléchargement: %v", downloadErr)
	}

	LogToFrontend("debug", fmt.Sprintf("Téléchargement terminé, %.2f MB téléchargés", float64(written)/(1024*1024)))

	// Le reste du code reste identique (décompression et restauration)
	LogToFrontend("debug", "Début de la décompression...")

	// Décompresse le tar.gz dans un dossier temporaire
	extractDir, err := os.MkdirTemp(tmpDir, "s3-restore-*")
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur création dossier temporaire: %v", err))
		return fmt.Errorf("erreur création dossier temporaire: %v", err)
	}
	defer os.RemoveAll(extractDir)
	LogToFrontend("debug", fmt.Sprintf("Extraction tar.gz dans: %s", extractDir))
	cmdTar := exec.Command("tar", "-xzf", tmpFilePath, "-C", extractDir)
	cmdTar.Stdout = os.Stdout
	cmdTar.Stderr = os.Stderr
	if err := cmdTar.Run(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur extraction tar.gz: %v", err))
		return fmt.Errorf("erreur extraction tar.gz: %v", err)
	}

	// On suppose que le dossier du bucket est à la racine de l'archive
	entries, err := os.ReadDir(extractDir)
	if err != nil || len(entries) == 0 {
		LogToFrontend("error", "Aucun dossier de bucket trouvé dans l'archive")
		return fmt.Errorf("aucun dossier de bucket trouvé dans l'archive")
	}
	bucketDir := entries[0].Name()
	bucketPath := extractDir + "/" + bucketDir
	LogToFrontend("debug", fmt.Sprintf("Bucket extrait: %s, chemin: %s", bucketDir, bucketPath))

	// Utilise les credentials locaux pour uploader dans le S3 local
	// Détermine le protocole à utiliser
	protocol := "http"
	if s3UseHttps {
		protocol = "https"
	}
	
	s3LocalCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(localCreds.AccessKey, localCreds.SecretKey, "")),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               fmt.Sprintf("%s://%s:%s", protocol, s3Host, s3Port),
					HostnameImmutable: true,
				}, nil
			}),
		),
	)
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur config S3 local: %v", err))
		return fmt.Errorf("erreur config S3 local: %v", err)
	}

	localClient := s3.NewFromConfig(s3LocalCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Vérifie si le bucket existe, sinon le crée
	_, err = localClient.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &bucketDir})
	if err != nil {
		LogToFrontend("debug", fmt.Sprintf("Bucket %s n'existe pas, création...", bucketDir))
		_, err = localClient.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &bucketDir})
		if err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur création bucket local: %v", err))
			return fmt.Errorf("erreur création bucket local: %v", err)
		}
	}
	LogToFrontend("debug", fmt.Sprintf("Début upload fichiers dans le bucket local: %s", bucketDir))

	dirEntries, err := os.ReadDir(bucketPath)
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur lecture du dossier bucket extrait: %v", err))
		return fmt.Errorf("erreur lecture du dossier bucket extrait: %v", err)
	}

	// Compte le nombre de fichiers pour afficher la progression
	totalFiles := 0
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			totalFiles++
		}
	}

	// Upload des fichiers avec barre de progression
	uploadedFiles := 0
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue // on ne gère que les fichiers à la racine
		}
		filePath := bucketPath + "/" + entry.Name()
		uploadedFiles++
		LogToFrontend("info", fmt.Sprintf("Upload fichier %d/%d: %s", uploadedFiles, totalFiles, entry.Name()))

		f, err := os.Open(filePath)
		if err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur ouverture fichier à restaurer: %v", err))
			return fmt.Errorf("erreur ouverture fichier à restaurer: %v", err)
		}

		fileInfo, err := f.Stat()
		if err == nil && fileInfo.Size() > 10*1024*1024 {
			LogToFrontend("debug", fmt.Sprintf("Fichier volumineux: %.2f MB", float64(fileInfo.Size())/(1024*1024)))
		}

		name := entry.Name()
		_, err = localClient.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketDir,
			Key:    &name,
			Body:   f,
		})
		f.Close()

		if err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur upload objet S3 local: %v", err))
			return fmt.Errorf("erreur upload objet S3 local: %v", err)
		}
	}

	// Supprime le fichier temporaire une fois terminé
	os.Remove(tmpFilePath)

	LogToFrontend("success", "Restauration S3 terminée avec succès.")
	return nil
}

func derefInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

type progressReaderWithLog struct {
	r          io.Reader
	total      int64
	read       int64
	last       int64
	lastUpdate time.Time
	lastLog    time.Time
}

func (p *progressReaderWithLog) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.read += int64(n)
		percent := int(float64(p.read) / float64(p.total) * 100)

		// Update progress at most every 500ms to avoid flooding the logs
		// Also update if percentage has changed, or at least every 5MB
		now := time.Now()
		percentChanged := percent != int(float64(p.last)/float64(p.total)*100)
		timePassed := now.Sub(p.lastUpdate) >= 500*time.Millisecond
		sizePassed := (p.read - p.last) >= 5*1024*1024 // 5MB

		if percentChanged && (timePassed || sizePassed) {
			mbRead := float64(p.read) / (1024 * 1024)
			mbTotal := float64(p.total) / (1024 * 1024)
			LogToFrontend("info", fmt.Sprintf("Téléchargement: %d%% (%.2f/%.2f MB)", percent, mbRead, mbTotal))
			p.last = p.read
			p.lastUpdate = now
		}
	}
	if err != nil && err != io.EOF {
		LogToFrontend("error", fmt.Sprintf("ERREUR lecture progressReaderWithLog: %v (lu jusqu'à présent: %d octets)", err, p.read))
	}
	return n, err
}

// DumpMongoDatabase crée un dump d'une base MongoDB
func DumpMongoDatabase(ctx context.Context, mongoHost, mongoPort, mongoUser, mongoPassword, database string) (string, error) {
	tmpDir, err := getUserTmpDir()
	if err != nil {
		return "", err
	}

	// Créer un fichier temporaire pour le dump
	tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("mongo-dump-%s-*.bson.gz", database))
	if err != nil {
		return "", fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Prépare la commande mongodump
	args := []string{
		"--gzip",
		"--archive=" + tmpFilePath,
		"--host", mongoHost,
		"--port", mongoPort,
		"--db", database,
	}
	if mongoUser != "" {
		args = append(args, "--username", mongoUser)
		// Ajouter la base d'authentification (par défaut admin)
		args = append(args, "--authenticationDatabase", "admin")
	}
	if mongoPassword != "" {
		args = append(args, "--password", mongoPassword)
	}

	LogToFrontend("info", fmt.Sprintf("Création du dump de la base %s...", database))
	cmd := exec.Command("mongodump", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(tmpFilePath)
		LogToFrontend("error", fmt.Sprintf("Erreur mongodump: %v", err))
		return "", fmt.Errorf("erreur mongodump: %v", err)
	}

	LogToFrontend("success", fmt.Sprintf("Dump de %s créé avec succès", database))
	return tmpFilePath, nil
}

// TransferMongoDatabase transfère une base de données entre deux serveurs MongoDB
func TransferMongoDatabase(ctx context.Context, sourceHost, sourcePort, sourceUser, sourcePassword, 
	destHost, destPort, destUser, destPassword, database string, dropExisting bool) error {
	
	LogToFrontend("info", fmt.Sprintf("Début du transfert de la base %s", database))
	
	// Étape 1: Créer le dump de la source
	dumpFile, err := DumpMongoDatabase(ctx, sourceHost, sourcePort, sourceUser, sourcePassword, database)
	if err != nil {
		return fmt.Errorf("erreur création dump: %v", err)
	}
	defer os.Remove(dumpFile)
	
	// Étape 2: Restaurer sur la destination
	args := []string{
		"--gzip",
		"--archive=" + dumpFile,
		"--host", destHost,
		"--port", destPort,
	}
	if destUser != "" {
		args = append(args, "--username", destUser)
		// Ajouter la base d'authentification (par défaut admin)
		args = append(args, "--authenticationDatabase", "admin")
	}
	if destPassword != "" {
		args = append(args, "--password", destPassword)
	}
	if dropExisting {
		args = append(args, "--drop")
	}
	
	LogToFrontend("info", fmt.Sprintf("Restauration de %s sur le serveur de destination...", database))
	cmd := exec.Command("mongorestore", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur mongorestore: %v", err))
		return fmt.Errorf("erreur mongorestore: %v", err)
	}
	
	LogToFrontend("success", fmt.Sprintf("Transfert de %s terminé avec succès", database))
	return nil
}

// ListMongoDatabases liste les bases de données disponibles sur un serveur MongoDB
func ListMongoDatabases(ctx context.Context, mongoHost, mongoPort, mongoUser, mongoPassword string) ([]string, error) {
	// Construire l'URI MongoDB
	var uri string
	if mongoUser != "" && mongoPassword != "" {
		// Ajouter authSource=admin pour l'authentification
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=admin", mongoUser, mongoPassword, mongoHost, mongoPort)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)
	}
	
	// Utiliser mongo shell pour lister les bases
	args := []string{
		uri,
		"--quiet",
		"--eval",
		"db.adminCommand('listDatabases').databases.forEach(function(d){print(d.name)})",
	}
	
	cmd := exec.Command("mongosh", args...)
	output, err := cmd.Output()
	if err != nil {
		// Fallback sur l'ancienne commande mongo si mongosh n'est pas disponible
		args[0] = uri
		cmd = exec.Command("mongo", args...)
		output, err = cmd.Output()
		if err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur listing databases: %v", err))
			return nil, fmt.Errorf("erreur listing databases: %v", err)
		}
	}
	
	// Parser la sortie pour obtenir la liste des bases
	lines := strings.Split(string(output), "\n")
	var databases []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "MongoDB") && !strings.Contains(line, "connecting") {
			databases = append(databases, line)
		}
	}
	
	return databases, nil
}

// RestorePostgresBackup télécharge un backup S3 et le restaure dans PostgreSQL
func RestorePostgresBackup(ctx context.Context, creds S3Credentials, s3Path string, pgHost, pgPort, pgUser, pgPassword, pgDatabase string) error {
	bucket := "backup-global"
	objectName := s3Path
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKey, creds.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("erreur chargement config AWS: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://" + S3BaseURL)
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
		LogToFrontend("info", fmt.Sprintf("Taille du backup à télécharger: %.2f MB", float64(totalSize)/(1024*1024)))
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
	tmpFile, err := os.CreateTemp(tmpDir, "postgres-backup-*.sql.gz")
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
	LogToFrontend("info", "Début du téléchargement du backup PostgreSQL...")
	_, err = io.Copy(tmpFile, progressReader)
	if err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}
	LogToFrontend("success", "Téléchargement du backup PostgreSQL terminé.")

	LogToFrontend("debug", fmt.Sprintf("pgHost=%s, pgPort=%s, pgUser=%s, pgDatabase=%s", pgHost, pgPort, pgUser, pgDatabase))

	// Définir PGPASSWORD dans l'environnement
	env := os.Environ()
	if pgPassword != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", pgPassword))
	}

	// Créer la base de données si elle n'existe pas
	LogToFrontend("info", fmt.Sprintf("Création de la base de données %s si elle n'existe pas...", pgDatabase))
	
	// D'abord vérifier si la base existe
	checkCmd := exec.Command("psql",
		"-h", pgHost,
		"-p", pgPort,
		"-U", pgUser,
		"-d", "postgres",
		"-t",
		"-c", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", pgDatabase))
	checkCmd.Env = env
	checkOutput, _ := checkCmd.Output()
	
	// Si la base n'existe pas, la créer
	if strings.TrimSpace(string(checkOutput)) == "" {
		createCmd := exec.Command("psql",
			"-h", pgHost,
			"-p", pgPort,
			"-U", pgUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("CREATE DATABASE %s", pgDatabase))
		createCmd.Env = env
		if err := createCmd.Run(); err != nil {
			LogToFrontend("error", fmt.Sprintf("Impossible de créer la base: %v", err))
			return fmt.Errorf("impossible de créer la base de données: %v", err)
		}
		LogToFrontend("success", fmt.Sprintf("Base de données %s créée avec succès", pgDatabase))
	}

	// Décompresser et restaurer avec pg_restore ou psql selon le format
	LogToFrontend("info", "Début de la restauration PostgreSQL...")
	
	// D'abord, décompresser le fichier pour déterminer son format
	cmd := exec.Command("gunzip", "-c", tmpFile.Name())
	cmd.Env = env
	unzippedData, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("erreur décompression: %v", err)
	}

	// Créer un fichier temporaire pour les données décompressées
	tmpUnzipped, err := os.CreateTemp(tmpDir, "postgres-backup-*.sql")
	if err != nil {
		return fmt.Errorf("erreur création fichier temporaire décompressé: %v", err)
	}
	defer os.Remove(tmpUnzipped.Name())
	
	if _, err := tmpUnzipped.Write(unzippedData); err != nil {
		tmpUnzipped.Close()
		return fmt.Errorf("erreur écriture fichier décompressé: %v", err)
	}
	tmpUnzipped.Close()

	// Vérifier si c'est un dump custom format ou SQL plain text
	fileCmd := exec.Command("file", tmpUnzipped.Name())
	fileOutput, _ := fileCmd.Output()
	fileType := string(fileOutput)

	var restoreCmd *exec.Cmd
	if strings.Contains(fileType, "PostgreSQL") && strings.Contains(fileType, "custom") {
		// Format custom, utiliser pg_restore
		restoreCmd = exec.Command("pg_restore",
			"-h", pgHost,
			"-p", pgPort,
			"-U", pgUser,
			"-d", pgDatabase,
			"--clean",
			"--if-exists",
			"--no-owner",
			"--no-privileges",
			tmpUnzipped.Name())
	} else {
		// Format SQL plain text, utiliser psql
		restoreCmd = exec.Command("psql",
			"-h", pgHost,
			"-p", pgPort,
			"-U", pgUser,
			"-d", pgDatabase,
			"-f", tmpUnzipped.Name())
	}

	restoreCmd.Env = env
	restoreCmd.Stdout = os.Stdout
	restoreCmd.Stderr = os.Stderr
	
	if err := restoreCmd.Run(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur restauration PostgreSQL: %v", err))
		return fmt.Errorf("erreur restauration PostgreSQL: %v", err)
	}

	LogToFrontend("success", "Restauration PostgreSQL terminée avec succès.")
	return nil
}

// DumpPostgresDatabase crée un dump d'une base PostgreSQL
func DumpPostgresDatabase(ctx context.Context, pgHost, pgPort, pgUser, pgPassword, database string) (string, error) {
	tmpDir, err := getUserTmpDir()
	if err != nil {
		return "", err
	}

	// Créer un fichier temporaire pour le dump
	tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("postgres-dump-%s-*.sql.gz", database))
	if err != nil {
		return "", fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Définir PGPASSWORD dans l'environnement
	env := os.Environ()
	if pgPassword != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", pgPassword))
	}

	LogToFrontend("info", fmt.Sprintf("Création du dump de la base %s...", database))
	
	// Utiliser pg_dump avec compression
	dumpCmd := exec.Command("pg_dump",
		"-h", pgHost,
		"-p", pgPort,
		"-U", pgUser,
		"-d", database,
		"--clean",
		"--if-exists",
		"--no-owner",
		"--no-privileges",
		"-Z", "9") // Compression maximale
	
	dumpCmd.Env = env
	dumpCmd.Stderr = os.Stderr

	// Récupérer la sortie de pg_dump
	dumpOutput, err := dumpCmd.Output()
	if err != nil {
		os.Remove(tmpFilePath)
		LogToFrontend("error", fmt.Sprintf("Erreur pg_dump: %v", err))
		return "", fmt.Errorf("erreur pg_dump: %v", err)
	}

	// Compresser avec gzip
	gzipCmd := exec.Command("gzip", "-9")
	gzipCmd.Stdin = strings.NewReader(string(dumpOutput))
	
	outFile, err := os.Create(tmpFilePath)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier de sortie: %v", err)
	}
	defer outFile.Close()
	
	gzipCmd.Stdout = outFile
	gzipCmd.Stderr = os.Stderr
	
	if err := gzipCmd.Run(); err != nil {
		os.Remove(tmpFilePath)
		LogToFrontend("error", fmt.Sprintf("Erreur compression gzip: %v", err))
		return "", fmt.Errorf("erreur compression gzip: %v", err)
	}

	LogToFrontend("success", fmt.Sprintf("Dump de %s créé avec succès", database))
	return tmpFilePath, nil
}

// TransferPostgresDatabase transfère une base de données entre deux serveurs PostgreSQL
func TransferPostgresDatabase(ctx context.Context, sourceHost, sourcePort, sourceUser, sourcePassword,
	destHost, destPort, destUser, destPassword, database string, dropExisting bool) error {

	LogToFrontend("info", fmt.Sprintf("Début du transfert de la base %s", database))

	// Étape 1: Créer le dump de la source
	dumpFile, err := DumpPostgresDatabase(ctx, sourceHost, sourcePort, sourceUser, sourcePassword, database)
	if err != nil {
		return fmt.Errorf("erreur création dump: %v", err)
	}
	defer os.Remove(dumpFile)

	// Étape 2: Décompresser le dump
	tmpDir, err := getUserTmpDir()
	if err != nil {
		return err
	}
	
	tmpUnzipped, err := os.CreateTemp(tmpDir, "postgres-transfer-*.sql")
	if err != nil {
		return fmt.Errorf("erreur création fichier temporaire: %v", err)
	}
	defer os.Remove(tmpUnzipped.Name())
	tmpUnzipped.Close()

	// Décompresser
	cmd := exec.Command("gunzip", "-c", dumpFile)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("erreur décompression: %v", err)
	}

	if err := os.WriteFile(tmpUnzipped.Name(), output, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier décompressé: %v", err)
	}

	// Étape 3: Gérer la base de données de destination
	env := os.Environ()
	if destPassword != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", destPassword))
	}

	if dropExisting {
		// Supprimer la base si elle existe
		LogToFrontend("info", fmt.Sprintf("Suppression de la base %s sur le serveur de destination si elle existe...", database))
		dropCmd := exec.Command("psql",
			"-h", destHost,
			"-p", destPort,
			"-U", destUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
		dropCmd.Env = env
		if err := dropCmd.Run(); err != nil {
			LogToFrontend("warn", fmt.Sprintf("Impossible de supprimer la base: %v", err))
		}
	}

	// Toujours créer la base de données si elle n'existe pas
	LogToFrontend("info", fmt.Sprintf("Création de la base %s si elle n'existe pas...", database))
	
	// D'abord vérifier si la base existe
	checkCmd := exec.Command("psql",
		"-h", destHost,
		"-p", destPort,
		"-U", destUser,
		"-d", "postgres",
		"-t",
		"-c", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", database))
	checkCmd.Env = env
	checkOutput, _ := checkCmd.Output()
	
	// Si la base n'existe pas, la créer
	if strings.TrimSpace(string(checkOutput)) == "" {
		createCmd := exec.Command("psql",
			"-h", destHost,
			"-p", destPort,
			"-U", destUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("CREATE DATABASE %s", database))
		createCmd.Env = env
		if err := createCmd.Run(); err != nil {
			LogToFrontend("error", fmt.Sprintf("Impossible de créer la base: %v", err))
			return fmt.Errorf("impossible de créer la base de données: %v", err)
		}
		LogToFrontend("success", fmt.Sprintf("Base de données %s créée avec succès", database))
	} else {
		LogToFrontend("info", fmt.Sprintf("La base de données %s existe déjà", database))
	}

	// Étape 4: Restaurer sur la destination
	LogToFrontend("info", fmt.Sprintf("Restauration de %s sur le serveur de destination...", database))
	restoreCmd := exec.Command("psql",
		"-h", destHost,
		"-p", destPort,
		"-U", destUser,
		"-d", database,
		"-f", tmpUnzipped.Name())
	
	restoreCmd.Env = env
	restoreCmd.Stdout = os.Stdout
	restoreCmd.Stderr = os.Stderr

	if err := restoreCmd.Run(); err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur restauration PostgreSQL: %v", err))
		return fmt.Errorf("erreur restauration PostgreSQL: %v", err)
	}

	LogToFrontend("success", fmt.Sprintf("Transfert de %s terminé avec succès", database))
	return nil
}

// ListPostgresDatabases liste les bases de données disponibles sur un serveur PostgreSQL
func ListPostgresDatabases(ctx context.Context, pgHost, pgPort, pgUser, pgPassword string) ([]string, error) {
	// Définir PGPASSWORD dans l'environnement
	env := os.Environ()
	if pgPassword != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", pgPassword))
	}

	// Utiliser psql pour lister les bases
	cmd := exec.Command("psql",
		"-h", pgHost,
		"-p", pgPort,
		"-U", pgUser,
		"-d", "postgres",
		"-t",
		"-c", "SELECT datname FROM pg_database WHERE datistemplate = false AND datname NOT IN ('postgres')")
	
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur listing databases PostgreSQL: %v", err))
		return nil, fmt.Errorf("erreur listing databases PostgreSQL: %v", err)
	}

	// Parser la sortie pour obtenir la liste des bases
	lines := strings.Split(string(output), "\n")
	var databases []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			databases = append(databases, line)
		}
	}

	return databases, nil
}
