package main

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunS3BucketUI lance l'interface pour naviguer et sélectionner des fichiers dans un bucket S3
func RunS3BucketUI(s3Manager *S3Manager) {
	app := tview.NewApplication()

	// Conteneur pour les fichiers sélectionnés
	selectedFiles := []string{}

	// Créer la racine de l'arborescence
	root := tview.NewTreeNode("Bucket: " + s3Manager.Bucket).SetColor(tcell.ColorGreen)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	// Charger les objets à la racine du bucket
	objects, prefixes, err := listObjects(s3Manager, "")
	if err != nil {
		log.Fatalf("Erreur lors du chargement des objets du bucket : %v", err)
	}
	addNodes(root, objects, prefixes)

	// Créer une vue texte pour afficher les fichiers sélectionnés
	selectedFilesList := tview.NewList()
	selectedFilesList.SetBorder(true)
	selectedFilesList.SetTitle("Fichiers sélectionnés")

	// Fonction pour mettre à jour la vue des fichiers sélectionnés
	updateSelectedFiles := func() {
		selectedFilesList.Clear() // Efface la liste existante
		for _, file := range selectedFiles {
			// Ajouter chaque fichier comme un élément dans la liste
			selectedFilesList.AddItem(file, "", 0, nil)
		}
	}
	// Configurer les actions de suppression dans selectedFilesList
	selectedFilesList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Supprimer le fichier sélectionné
		if index >= 0 && index < len(selectedFiles) {
			selectedFiles = append(selectedFiles[:index], selectedFiles[index+1:]...)
			updateSelectedFiles()
		}
	})

	// En-tête
	header := tview.NewTextView().
		SetText("Aidalinfo devcli 🚀 - Backup management").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Pied de page
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]↑/↓ : Naviguer | Entrée : Sélectionner | Échap : Quitter").
		SetTextAlign(tview.AlignCenter)

	// Grille pour organiser l'interface
	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(90, 70).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(tree, 1, 0, 1, 1, 0, 0, true).
		AddItem(selectedFilesList, 1, 1, 1, 1, 0, 0, false).
		AddItem(footer, 2, 0, 1, 2, 0, 0, false)

	// Configurer l'action lors de la sélection dans l'arborescence
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		// Si le nœud sélectionné n'a pas de référence, ne rien faire
		if node.GetReference() == nil {
			return
		}

		// Récupérer le préfixe (chemin du dossier ou fichier)
		prefix := node.GetReference().(string)

		// Si c'est un fichier, demander une confirmation pour le sélectionner
		if !strings.HasSuffix(prefix, "/") {
			confirmation := tview.NewModal().
				SetText(fmt.Sprintf("Voulez-vous sélectionner le fichier suivant ?\n\n%s", prefix)).
				AddButtons([]string{"Oui", "Non"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Oui" {
						selectedFiles = append(selectedFiles, prefix)
						updateSelectedFiles()
					}
					app.SetRoot(grid, true).SetFocus(tree)
				})
			app.SetRoot(confirmation, true).SetFocus(confirmation)
			return
		}

		// Si c'est un dossier, charger son contenu
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}

		objects, prefixes, err := listObjects(s3Manager, prefix)
		if err != nil {
			log.Printf("Erreur lors du chargement du contenu du dossier %s : %v", prefix, err)
			return
		}
		addNodes(node, objects, prefixes)
		node.SetExpanded(true)
	})

	// Configurer la touche Tab pour naviguer entre les deux éléments
	currentFocus := 0
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape: // Quitter l'application
			app.Stop()
			return nil
		case tcell.KeyTab: // Basculer vers le prochain élément
			currentFocus = (currentFocus + 1) % 2
			if currentFocus == 0 {
				app.SetFocus(tree)
			} else {
				app.SetFocus(selectedFilesList)
			}
			return nil
		case tcell.KeyBacktab: // Basculer vers l'élément précédent
			currentFocus = (currentFocus + 1) % 2
			if currentFocus == 0 {
				app.SetFocus(tree)
			} else {
				app.SetFocus(selectedFilesList)
			}
			return nil
		}
		return event
	})

	// Lancer l'application
	if err := app.SetRoot(grid, true).SetFocus(tree).Run(); err != nil {
		log.Fatalf("Erreur lors du lancement de l'application : %v", err)
	}
}

// listObjects liste les objets et dossiers dans un bucket S3 avec un préfixe donné
func listObjects(s3Manager *S3Manager, prefix string) ([]types.Object, []string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    &s3Manager.Bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"), // Délimiteur pour séparer les dossiers
	}

	output, err := s3Manager.Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, nil, err
	}

	// Convertir les CommonPrefixes en []string
	var prefixes []string
	for _, cp := range output.CommonPrefixes {
		if cp.Prefix != nil {
			prefixes = append(prefixes, *cp.Prefix)
		}
	}

	return output.Contents, prefixes, nil
}

// addNodes ajoute les fichiers et dossiers à un nœud parent
func addNodes(parentNode *tview.TreeNode, objects []types.Object, prefixes []string) {
	// Ajouter les dossiers
	for _, prefix := range prefixes {
		name := path.Base(strings.TrimSuffix(prefix, "/"))
		node := tview.NewTreeNode(name + "/").
			SetReference(prefix).
			SetSelectable(true).
			SetColor(tcell.ColorBlue)
		parentNode.AddChild(node)
	}

	// Ajouter les fichiers
	for _, obj := range objects {
		if obj.Key == nil {
			continue
		}
		name := path.Base(*obj.Key)
		node := tview.NewTreeNode(name).
			SetReference(*obj.Key).
			SetSelectable(true).
			SetColor(tcell.ColorWhite)
		parentNode.AddChild(node)
	}
}
