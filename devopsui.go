package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func showSubmodulesTagsDetails(app *tview.Application, selectedSubmodules []string, grid *tview.Grid, list *tview.List) {
	detailView := tview.NewFlex().SetDirection(tview.FlexRow)

	// Header
	header := tview.NewTextView().
		SetText("Aidalinfo devcli 🚀 - Détails des tags").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	detailView.AddItem(header, 3, 0, false)

	// Créer un Flex pour contenir tous les sous-modules
	modulesContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	// Créer un tableau pour stocker toutes les listes de tags
	focusableElements := make([]tview.Primitive, 0)
	currentFocusIndex := 0

	// Pour chaque sous-module sélectionné
	for _, submodule := range selectedSubmodules {
		currentBranch := getCurrentBranch(submodule)
		submodulePath := submodule // Capture la valeur actuelle

		// Vue pour afficher le nom du sous-module
		submoduleNameView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[green]Sous-module : [white]%s", submodule)).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		// Vue pour afficher la branche actuelle
		branchView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[yellow]Branche actuelle : [white]%s", currentBranch)).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		// Bouton pour créer un nouveau tag
		newTagButton := tview.NewButton("Créer un nouveau tag").
			SetSelectedFunc(func() {
				showCreateTagModal(app, submodulePath, detailView)
			})
		newTagButton.SetBorder(true).SetTitle("Créer un tag")
		// Vue pour afficher les modifications en attente
		modificationsView := tview.NewTextView().
			SetDynamicColors(true).
			SetWrap(true)
		modifications, err := getWaitingChanges(submodulePath)
		if err != nil {
			modificationsView.SetText(fmt.Sprintf("[red]Erreur : [white]%s", err))
		} else if modifications == "" {
			modificationsView.SetText("[green]Aucune modification en attente.")
		} else {
			modificationsView.SetText(fmt.Sprintf("[yellow]Modifications en attente :\n[white]%s", modifications))
		}
		// Liste des tags stables `v*`
		tagsVList := tview.NewList()
		// Liste des tags beta `rc-*`
		tagsRCList := tview.NewList()

		// Ajouter le bouton et les listes aux éléments focalisables
		focusableElements = append(focusableElements, newTagButton, tagsVList, tagsRCList)

		// Récupérer les tags
		vTags, rcTags, err := getLastTags(submodulePath)
		if err != nil {
			// Afficher une modale d'erreur
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Erreur lors de la récupération des tags pour %s : %s", submodulePath, err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.SetRoot(detailView, true)
				})
			app.SetRoot(modal, true)
			return
		}

		// Ajouter les tags stables `v*` à la liste correspondante
		for _, tag := range vTags {
			tagsVList.AddItem(tag, "", 0, nil)
		}
		tagsVList.SetBorder(true).SetTitle("Tags stables (v*)")

		// Ajouter les tags beta `rc-*` à la liste correspondante
		for _, tag := range rcTags {
			tagsRCList.AddItem(tag, "", 0, nil)
		}
		tagsRCList.SetBorder(true).SetTitle("Tags beta (rc-*)")

		// Vue combinée pour ce sous-module
		moduleView := tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(submoduleNameView, 2, 0, false). // Nom du sous-module en haut
				AddItem(branchView, 2, 0, false).       // Branche actuelle
				AddItem(newTagButton, 2, 0, true). // Bouton pour créer un nouveau tag
				AddItem(modificationsView, 4, 0, false), // Modifications en attente
				0, 1, false).
			AddItem(tagsVList, 0, 2, true).  // Tags stables au milieu
			AddItem(tagsRCList, 0, 2, true) // Tags beta à droite

		// Ajouter ce module dans le conteneur principal
		modulesContainer.AddItem(moduleView, 0, 1, true).
			AddItem(tview.NewBox().SetBackgroundColor(tcell.ColorGray), 1, 0, false) // Séparateur visuel
	}

	detailView.AddItem(modulesContainer, 0, 1, true)

	// Gestion du focus et des touches
	detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		totalElements := len(focusableElements)

		switch event.Key() {
		case tcell.KeyEscape:
			app.SetRoot(grid, true).SetFocus(list)
			return nil
		case tcell.KeyTab:
			// Passer au prochain élément
			currentFocusIndex = (currentFocusIndex + 1) % totalElements
			app.SetFocus(focusableElements[currentFocusIndex])
			return nil
		case tcell.KeyBacktab:
			// Passer à l'élément précédent
			currentFocusIndex--
			if currentFocusIndex < 0 {
				currentFocusIndex = totalElements - 1
			}
			app.SetFocus(focusableElements[currentFocusIndex])
			return nil
		}
		return event
	})

	// Définir le focus initial sur le premier élément focalisable
	if len(focusableElements) > 0 {
		app.SetFocus(focusableElements[0])
	}

	// Définir la vue de détails comme racine
	app.SetRoot(detailView, true)
}


func showCreateTagModal(app *tview.Application, repoPath string, previousView tview.Primitive) {
	// Créer un formulaire
	form := tview.NewForm()

	// Ajouter des champs d'entrée et des boutons
	form.AddInputField("Version", "", 20, nil, nil).
		AddInputField("Message", "", 50, nil, nil).
		AddButton("Créer", func() {
			// Récupérer les données saisies dans le formulaire
			version := form.GetFormItemByLabel("Version").(*tview.InputField).GetText()
			message := form.GetFormItemByLabel("Message").(*tview.InputField).GetText()

			// Tenter de créer le tag
			err := createTag(repoPath, version, message)
			if err != nil {
				// Afficher une modale d'erreur en cas d'échec
				errorModal := tview.NewModal().
					SetText(fmt.Sprintf("Erreur lors de la création du tag : %s", err)).
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						// Retourner à la vue précédente
						app.SetRoot(previousView, true)
					})
				app.SetRoot(errorModal, true)
			} else {
				// Afficher une modale de succès
				successModal := tview.NewModal().
					SetText(fmt.Sprintf("Tag %s créé avec succès.", version)).
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						// Retourner à la vue précédente
						app.SetRoot(previousView, true)
					})
				app.SetRoot(successModal, true)
			}
		}).
		AddButton("Annuler", func() {
			// Retourner à la vue précédente si annulé
			app.SetRoot(previousView, true)
		})

	// Configurer le formulaire
	form.SetBorder(true).
		SetTitle("Créer un nouveau tag").
		SetTitleAlign(tview.AlignCenter)

	// Afficher le formulaire dans l'application
	app.SetRoot(form, true).SetFocus(form)
}




func RunDevOpsUI(submodules []string, submoduleNames []string) {
	app := tview.NewApplication()
	selectedSubmodules := []string{}

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
                    showSubmodulesTagsDetails(app, selectedSubmodules, grid, list)
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