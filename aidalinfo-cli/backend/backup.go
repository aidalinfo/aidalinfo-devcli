package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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

// ListBackupsWithCreds liste les backups S3 avec credentials fournis (bucket privé, signature S3 via AWS SDK)
func ListBackupsWithCreds(ctx context.Context, creds S3Credentials, s3Dir string) ([]string, error) {
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
	var files []string
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	paginator := s3.NewListObjectsV2Paginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du listing S3: %v", err)
		}
		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, "/") {
				files = append(files, *obj.Key)
			}
		}
	}
	return files, nil
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
