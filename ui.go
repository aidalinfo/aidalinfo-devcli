package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Ajouter cette fonction qui affiche les détails des sous-modules
func showSubmodulesDetails(app *tview.Application, selectedSubmodules []string, grid *tview.Grid, list *tview.List) {
	detailView := tview.NewFlex().SetDirection(tview.FlexRow)

	header := tview.NewTextView().
		SetText("Aidalinfo devcli 🚀 - Détails des sous-modules").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	detailView.AddItem(header, 3, 0, false)

	// Créer un Flex pour contenir tous les sous-modules
	modulesContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	// Créer un tableau pour stocker toutes les listes de branches
	branchesLists := make([]*tview.List, 0)
	currentFocusIndex := 0

	// Pour chaque sous-module sélectionné
	for _, submodule := range selectedSubmodules {
		currentBranch := getCurrentBranch(submodule)
		submodulePath := submodule // Capture la valeur actuelle

		branchView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[yellow]Branche actuelle : [white]%s", currentBranch)).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		branchesList := tview.NewList()
		branchesLists = append(branchesLists, branchesList)
		branches := getBranches(submodulePath)

		for _, branch := range branches {
			if branch == "" {
				continue
			}

			branchName := branch
			currentBranchName := currentBranch

			branchesList.AddItem(branchName, "", 0, func() {
				diffSummary, err := getDiffSummary(currentBranchName, branchName, submodulePath)
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

				modal := tview.NewModal().
					SetText(fmt.Sprintf("Êtes-vous sûr de vouloir merger %s dans %s ?\n\nRésumé des différences :\n%s",
						branchName, currentBranchName, diffSummary)).
					AddButtons([]string{"Oui", "Non"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonLabel == "Oui" {
							err := createMerge(currentBranchName, branchName, submodulePath)
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
									SetText(fmt.Sprintf("Le merge de %s dans %s a réussi.",
										branchName, currentBranchName)).
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
			})
		}
		branchesList.SetBorder(true).SetTitle("Branches disponibles")

		moduleView := tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(branchView, 0, 1, false).
			AddItem(branchesList, 0, 3, true)

		moduleTitle := tview.NewTextView().
			SetText(fmt.Sprintf("[yellow]%s[white]", submodule)).
			SetDynamicColors(true)

		modulesContainer.AddItem(moduleTitle, 1, 0, false)
		modulesContainer.AddItem(moduleView, 0, 1, true)
		modulesContainer.AddItem(tview.NewBox().SetBackgroundColor(tcell.ColorGray), 1, 0, false)
	}

	detailView.AddItem(modulesContainer, 0, 1, true)

	// Gestion du focus et des touches
	detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.SetRoot(grid, true).SetFocus(list)
			return nil
		case tcell.KeyTab:
			// Passer au prochain sous-module
			currentFocusIndex = (currentFocusIndex + 1) % len(branchesLists)
			app.SetFocus(branchesLists[currentFocusIndex])
			return nil
		case tcell.KeyBacktab:
			// Passer au sous-module précédent
			currentFocusIndex--
			if currentFocusIndex < 0 {
				currentFocusIndex = len(branchesLists) - 1
			}
			app.SetFocus(branchesLists[currentFocusIndex])
			return nil
		}
		return event
	})

	// Définir le focus initial sur la première liste de branches
	if len(branchesLists) > 0 {
		app.SetFocus(branchesLists[0])
	}

	app.SetRoot(detailView, true)
}

func RunUI(submodules []string, submoduleNames []string) {
	app := tview.NewApplication()
	selectedSubmodules := []string{}

	// Liste des sous-modules
	list := tview.NewList()

	// Récupérer les commits
	commits, err := getLastCommits(submodules)
	if err != nil {
		panic(err)
	}

	// Créer la vue des commits
	commitsView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetScrollable(true)

	commitsView.SetBorder(true).SetTitle("Derniers commits")

	// Formater et afficher les commits
	content := ""
	for _, commit := range commits {
		content += fmt.Sprintf(
			"[yellow]%s[white]\n"+
				"[blue]%s[white] - %s\n"+
				"%s\n"+
				"-------------------\n",
			commit.Submodule,
			commit.Date[:16],
			commit.Author,
			commit.Message,
		)
	}
	commitsView.SetText(content)

	// SetInputCapture pour commitsView
	commitsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(list)
			return nil
		}
		return event
	})

	// Barre latérale
	sideBar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	sideBar.SetBorder(true).
		SetTitle("Séléctionné(s)")

	// Fonction pour mettre à jour la barre latérale
	updateSidebar := func() {
		content := ""
		for _, module := range selectedSubmodules {
			content += fmt.Sprintf("%s\n", module)
		}
		sideBar.SetText(content)
	}

	// Remplir la liste des sous-modules
	for index, submoduleName := range submoduleNames {
		idx := index
		list.AddItem(submoduleName, "", 0, func() {
			selected := false
			for i, module := range selectedSubmodules {
				if module == submodules[idx] {
					selectedSubmodules = append(selectedSubmodules[:i], selectedSubmodules[i+1:]...)
					selected = true
					break
				}
			}
			if !selected {
				selectedSubmodules = append(selectedSubmodules, submodules[idx])
			}
			updateSidebar()
		})
	}
	list.SetBorder(true).SetTitle("Submodules")

	// Footer
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]i: chercher | n: suivant | espace: selectionner | tab: naviguer | ↑↓: scroll").
		SetTextAlign(tview.AlignLeft)

	// Grid Layout
	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(40, 0, 30).
		SetBorders(true).
		AddItem(tview.NewTextView().
			SetText("Aidalinfo devcli 🚀").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true), 0, 0, 1, 3, 0, 0, false).
		AddItem(commitsView, 1, 0, 1, 1, 0, 100, false).
		AddItem(list, 1, 1, 1, 1, 0, 100, true).
		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false).
		AddItem(footer, 2, 0, 1, 3, 0, 0, false)

	// Gestion des touches pour la liste
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case ' ':
				index := list.GetCurrentItem()
				selected := false
				for i, module := range selectedSubmodules {
					if module == submodules[index] {
						selectedSubmodules = append(selectedSubmodules[:i], selectedSubmodules[i+1:]...)
						selected = true
						break
					}
				}
				if !selected {
					selectedSubmodules = append(selectedSubmodules, submodules[index])
				}
				updateSidebar()
				return nil
			case 'n':
				if len(selectedSubmodules) > 0 {
					showSubmodulesDetails(app, selectedSubmodules, grid, list)
				}
				return nil
			}
		} else if event.Key() == tcell.KeyTab {
			app.SetFocus(commitsView)
			return nil
		}
		return event
	})

	// Gestion du focus global
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.GetFocus() == commitsView {
			switch event.Key() {
			case tcell.KeyUp:
				row, _ := commitsView.GetScrollOffset()
				commitsView.ScrollTo(row-1, 0)
				return nil
			case tcell.KeyDown:
				row, _ := commitsView.GetScrollOffset()
				commitsView.ScrollTo(row+1, 0)
				return nil
			case tcell.KeyPgUp:
				row, _ := commitsView.GetScrollOffset()
				commitsView.ScrollTo(row-10, 0)
				return nil
			case tcell.KeyPgDn:
				row, _ := commitsView.GetScrollOffset()
				commitsView.ScrollTo(row+10, 0)
				return nil
			}
		}
		return event
	})

	// Lancer l'application
	if err := app.SetRoot(grid, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
