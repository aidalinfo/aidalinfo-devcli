package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunUI démarre l'interface utilisateur
func RunUI(submodules []string, submoduleNames []string) {
	// Créer une nouvelle application tview
	app := tview.NewApplication()

	// Définir la vue principale
	var mainView tview.Primitive

	// Fonction pour afficher une vue détaillée d'un sous-module
	showSubmoduleView := func(submoduleName string, submodulePath string) {
		// Vue pour la branche actuelle
		branchView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[yellow]Branche actuelle : [white]%s", getCurrentBranch(submodulePath))).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		// Vue pour la liste des branches disponibles
		branchesList := tview.NewList()
		branches := getBranches(submodulePath)
		for _, branch := range branches {
			if branch == "" {
				continue
			}
			branchesList.AddItem(branch, "", 0, nil)
		}
		branchesList.SetBorder(true).SetTitle("Branches disponibles")

		// Vue détaillée avec disposition
		detailView := tview.NewFlex().
			AddItem(branchView, 0, 1, false). // Partie gauche : branche actuelle
			AddItem(branchesList, 0, 2, true) // Partie droite : liste des branches

		// Gestion de la touche Échap pour revenir à la vue principale
		app.SetRoot(detailView, true).SetFocus(detailView)
		detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				app.SetRoot(mainView, true).SetFocus(mainView)
				return nil
			}
			return event
		})
	}

	// Créer une liste des sous-modules
	list := tview.NewList()
	for i, submoduleName := range submoduleNames {
		// Ajouter chaque sous-module avec le chemin complet
		list.AddItem(submoduleName, "", 0, func(name, path string) func() {
			return func() {
				showSubmoduleView(name, path)
			}
		}(submoduleName, submodules[i]))
	}

	// Vue principale avec un titre
	mainView = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetText("Liste des sous-modules").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true), 3, 1, false).
		AddItem(list, 0, 1, true)

	// Définir la vue principale comme racine
	if err := app.SetRoot(mainView, true).Run(); err != nil {
		panic(err)
	}
}
