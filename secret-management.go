package main

import (
	"context"
	"fmt"
	"os"

	infisical "github.com/infisical/go-sdk"
)

func getSecret(secretName string, environment string) (string, error) {
	// Récupération des variables d'environnement
	if environment == "" {
		environment = "prod"
	}
	infisicalURL := os.Getenv("INFISICAL_URL")
	accessToken := os.Getenv("INFISICAL_API_KEY")

	if infisicalURL == "" {
		infisicalURL = "https://app.infisical.com"
	}

	if accessToken == "" {
		return "", fmt.Errorf("l'environnement INFISICAL_API_KEY est manquant")
	}

	// Initialisation du client Infisical
	client := infisical.NewInfisicalClient(context.Background(), infisical.Config{
		SiteUrl:          infisicalURL,
		AutoTokenRefresh: false,
	})

	// Authentification avec l'API Key
	client.Auth().SetAccessToken(accessToken)

	// Récupération du secret
	secret, err := client.Secrets().Retrieve(infisical.RetrieveSecretOptions{
		SecretKey:   secretName,
		Environment: environment,
		ProjectID:   os.Getenv("INFISICAL_PROJECT_ID"),
		SecretPath:  "/",
	})
	if err != nil {
		return "", fmt.Errorf("échec de la récupération du secret: %v", err)
	}

	return secret.SecretValue, nil
}
