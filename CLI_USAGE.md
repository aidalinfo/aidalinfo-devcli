# Aidalinfo CLI - Documentation des commandes

## Installation
```bash
go build -o aidalinfo-cli
```

## Utilisation

### Mode GUI (par défaut)
Lancez simplement l'application sans arguments :
```bash
./aidalinfo-cli
```

### Mode CLI
Toutes les commandes sont disponibles avec des arguments :

#### Installation des submodules
```bash
# Installation simple des submodules
./aidalinfo-cli install

# Installation avec branches spécifiques
./aidalinfo-cli install --branch "develop staging"

# Installation avec NPM
./aidalinfo-cli install --npm

# Spécifier un chemin de projet
./aidalinfo-cli install --path /chemin/vers/projet
```

#### Installation NPM
```bash
# Installer les dépendances NPM
./aidalinfo-cli npm

# Dans un projet spécifique
./aidalinfo-cli npm --path /chemin/vers/projet
```

#### Installation complète
```bash
# Installation complète (submodules + NPM)
./aidalinfo-cli full

# Avec chemin spécifique
./aidalinfo-cli full --path /chemin/vers/projet
```

#### Mise à jour
```bash
# Mettre à jour les dépendances NPM
./aidalinfo-cli update-npm

# Mettre à jour les sous-modules (git pull)
./aidalinfo-cli update-git
```

#### Gestion des tags
```bash
# Créer un tag
./aidalinfo-cli tag --name "v1.0.0" --message "Version 1.0.0"

# Créer un tag pour un submodule spécifique
./aidalinfo-cli tag --name "v1.0.0" --message "Version 1.0.0" --submodule "frontend"
```

#### Lister les submodules
```bash
# Lister tous les submodules
./aidalinfo-cli list

# Dans un projet spécifique
./aidalinfo-cli list --path /chemin/vers/projet
```

#### Backup et restauration
```bash
# Backup vers S3
./aidalinfo-cli backup --type s3 --s3-bucket "mon-bucket"

# Backup local
./aidalinfo-cli backup --type local --local-path "/chemin/vers/backup"

# Restauration depuis S3
./aidalinfo-cli backup --restore --type s3 --s3-bucket "mon-bucket"

# Restauration locale
./aidalinfo-cli backup --restore --type local --local-path "/chemin/vers/backup.tar.gz"
```

#### Autres commandes
```bash
# Afficher la version
./aidalinfo-cli version

# Aide
./aidalinfo-cli --help

# Aide pour une commande spécifique
./aidalinfo-cli install --help
```

## Options globales

- `--path` : Spécifie le chemin du projet (par défaut : répertoire courant)

## Exemples d'utilisation

### Workflow typique de développement

1. Cloner un projet et installer les dépendances :
```bash
git clone https://github.com/monprojet/repo.git
cd repo
aidalinfo-cli full
```

2. Travailler sur une branche spécifique :
```bash
aidalinfo-cli install --branch "feature-xyz"
```

3. Mettre à jour le projet :
```bash
aidalinfo-cli update-git
aidalinfo-cli update-npm
```

4. Créer une release :
```bash
aidalinfo-cli tag --name "v2.0.0" --message "Release 2.0.0"
```

5. Sauvegarder le projet :
```bash
aidalinfo-cli backup --type local --local-path ~/backups
```