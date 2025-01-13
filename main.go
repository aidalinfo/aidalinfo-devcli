package main

import (
    "github.com/rivo/tview"
    "fmt"
    // "regexp"
)

func main() {
    // Appelle listSubmodule pour le répertoire courant
    submodules, err := listSubmodule("/home/killian/dev/aidalinfo/PROJET-pulse-myIT")
    if err != nil {
        fmt.Println("Erreur :", err)
        return
    }
    submoduleNames, err := cleanSubmodule(submodules)
    if err != nil {
        fmt.Println("Erreur :", err)
        return
    }

    // re := regexp.MustCompile(`[^/]+$`)

    // // Extraire uniquement la dernière partie de chaque chemin
    // var submoduleNames []string
    // for _, submodule := range submodules {
    //     // Utilise la regex pour trouver la partie après le dernier "/"
    //     matches := re.FindStringSubmatch(submodule)
    //     if len(matches) > 0 {
    //         submoduleNames = append(submoduleNames, matches[0])
    //     }
    // }

    // Affiche les sous-modules trouvés
    fmt.Println("Sous-modules trouvés :")
    for _, submodule := range submodules {
        fmt.Println(submodule)
    }

    // Créer une nouvelle application
    app := tview.NewApplication()

    // Créer un texte pour le titre
    title := tview.NewTextView().
        SetText("Bienvenue dans aidalinfo-devcli").
        SetTextAlign(tview.AlignCenter).
        SetDynamicColors(true).
        SetTextColor(tview.Styles.TitleColor)

    // Crée une table pour afficher les sous-modules
    table := tview.NewTable().SetSelectable(true, false)
    selected := make(map[int]bool) // Stocke les indices des sous-modules sélectionnés

    // Ajoute les sous-modules à la table avec des cases à cocher
    for i, name := range submoduleNames {
        // Colonne 0 : Case à cocher (initialement vide)
        table.SetCell(i, 0, tview.NewTableCell("[ ]").
            SetTextColor(tview.Styles.PrimitiveBackgroundColor))
        // Colonne 1 : Nom du sous-module
        table.SetCell(i, 1, tview.NewTableCell(name).
            SetTextColor(tview.Styles.PrimaryTextColor))
    }

    // Gère la sélection et les cases à cocher
    table.SetSelectedFunc(func(row, column int) {
        if column == 0 {
            // Inverse l'état de la case à cocher
            if selected[row] {
                table.GetCell(row, 0).SetText("[ ]")
                delete(selected, row)
            } else {
                table.GetCell(row, 0).SetText("[X]")
                selected[row] = true
            }
            app.Draw() // Redessine l'interface
        }
    })

    // Disposition globale
    flex := tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(title, 3, 1, false).
        AddItem(table, 0, 1, true)

    // Lancer l'application
    if err := app.SetRoot(flex, true).Run(); err != nil {
        panic(err)
    }
}