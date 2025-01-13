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
		currentBranch := getCurrentBranch(submodulePath)
		branchView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[yellow]Branche actuelle : [white]%s", currentBranch)).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		// Vue pour la liste des branches disponibles
		branchesList := tview.NewList()
		branches := getBranches(submodulePath)

		var detailView *tview.Flex // Déclare detailView en dehors des closures

		for _, branch := range branches {
			if branch == "" {
				continue
			}
			branchesList.AddItem(branch, "", 0, func(targetBranch string) func() {
				return func() {
					// Obtenir le résumé des différences
					diffSummary, err := getDiffSummary(currentBranch, targetBranch, submodulePath)
					if err != nil {
						errorModal := tview.NewModal().
							SetText(fmt.Sprintf("Erreur : %s", err)).
							AddButtons([]string{"OK"}).
							SetDoneFunc(func(int, string) {
								app.SetRoot(detailView, true)
							})
						app.SetRoot(errorModal, true)
						return
					}
					// Demander confirmation pour le merge
					modal := tview.NewModal().
						SetText(fmt.Sprintf("Êtes-vous sûr de vouloir merger %s dans %s ? \n\nRésumé des différences :\n%s", targetBranch, currentBranch, diffSummary)).
						AddButtons([]string{"Oui", "Non"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "Oui" {
								// Effectuer le merge
								err := createMerge(currentBranch, targetBranch, submodulePath)
								if err != nil {
									errorModal := tview.NewModal().
										SetText(fmt.Sprintf("Erreur : %s", err)).
										AddButtons([]string{"OK"}).
										SetDoneFunc(func(int, string) {
											app.SetRoot(detailView, true)
										})
									app.SetRoot(errorModal, true)
								} else {
									successModal := tview.NewModal().
										SetText(fmt.Sprintf("Le merge de %s dans %s a réussi.", targetBranch, currentBranch)).
										AddButtons([]string{"OK"}).
										SetDoneFunc(func(int, string) {
											app.SetRoot(detailView, true)
										})
									app.SetRoot(successModal, true)
								}
							} else {
								app.SetRoot(detailView, true)
							}
						})
					app.SetRoot(modal, true)
				}
			}(branch))
		}
		branchesList.SetBorder(true).SetTitle("Branches disponibles")

		// Vue détaillée avec disposition
		detailView = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewTextView().
				SetText(fmt.Sprintf("[green]Aidalinfo devcli 🚀[/green]\n\n[submodule: %s]", submoduleName)).
				SetTextAlign(tview.AlignCenter).
				SetDynamicColors(true), 3, 1, false).
			AddItem(tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(branchView, 0, 1, false).
				AddItem(branchesList, 0, 3, true), 0, 1, true)

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
	list.SetBorder(false) // Supprime uniquement la bordure

	// Barre d'état dynamique
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]main > Sélectionnez un sous-module").
		SetTextAlign(tview.AlignLeft)

	// Gestion des événements pour mettre à jour la barre d'état
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		statusBar.SetText(fmt.Sprintf("[green]main > [blue]%s", mainText))
	})

	// Vue principale avec un titre
	mainView = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetText("Aidalinfo devcli 🚀").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true), 3, 1, false).
		AddItem(list, 0, 1, true).
		AddItem(statusBar, 1, 1, false)

	// Définir la vue principale comme racine
	if err := app.SetRoot(mainView, true).Run(); err != nil {
		panic(err)
	}
}
