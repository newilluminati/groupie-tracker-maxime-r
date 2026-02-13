package gui

import (
	"fmt"
	"sync"

	"groupie-tracker/api"
	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// app.go - le coeur de l'application Fyne
// c'est ici qu'on gere la fenetre principale et la navigation entre les pages

// AppGroupie - la struct principale de l'app, elle contient tout ce dont on a besoin
type AppGroupie struct {
	app              fyne.App
	fenetre          fyne.Window
	artistes         []models.Artiste
	locationsData    models.IndexLocations
	contenuPrinc     *fyne.Container // le container principal ou on met les pages
	pageAccueil      fyne.CanvasObject
	cacheImages      map[int][]byte // cache des images telecharges
	cacheImagesMu    sync.RWMutex
	favoris          map[int]bool // les favoris de l'utilisateur
	favorisMu        sync.RWMutex
	barreRecherche   *EntryRecherche // la barre de recherche (pour les raccourcis)
	onRefreshAccueil func()          // callback pour rafraichir la page d'accueil
}

// LancerApp - point d'entrée de l'interface graphique
// on charge les données et on lance la fenetre
func LancerApp() {
	// on cree l'app Fyne
	monApp := app.NewWithID("com.groupie.tracker")
	monApp.Settings().SetTheme(theme.DarkTheme())

	fenetre := monApp.NewWindow("Groupie Tracker 🎵")
	fenetre.Resize(fyne.NewSize(1200, 800))
	fenetre.CenterOnScreen()

	// ptit label de chargement pendant qu'on fetch les donnees
	labelChargement := widget.NewLabel("⏳ Chargement des artistes...")
	labelChargement.Alignment = fyne.TextAlignCenter
	fenetre.SetContent(container.NewCenter(labelChargement))
	fenetre.Show()

	// on charge les donnees en arriere-plan pour pas bloquer la fenetre
	go func() {
		artistes, err := api.RecupererArtistes()
		if err != nil {
			labelChargement.SetText(fmt.Sprintf("❌ Erreur: %v", err))
			return
		}

		locData, err := api.RecupererToutesLocations()
		if err != nil {
			// c'est pas grave si on a pas les locations, on continue quand meme
			fmt.Println("Warning: impossible de charger les locations:", err)
		}

		appGrp := &AppGroupie{
			app:           monApp,
			fenetre:       fenetre,
			artistes:      artistes,
			locationsData: locData,
			cacheImages:   make(map[int][]byte),
			favoris:       make(map[int]bool),
		}

		// on setup les raccourcis clavier
		appGrp.setupRaccourcis()

		// on cree la page d'accueil et on l'affiche
		appGrp.afficherAccueil()
	}()

	monApp.Run()
}

// setupRaccourcis - configure les raccourcis clavier globaux
// Ctrl+F -> focus sur la recherche, Escape -> retour accueil
func (a *AppGroupie) setupRaccourcis() {
	// Ctrl+F pour la recherche
	ctrlF := &desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}
	a.fenetre.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		// on focus la barre de recherche si elle existe
		if a.barreRecherche != nil {
			a.fenetre.Canvas().Focus(a.barreRecherche)
		}
	})

	// Ctrl+H pour retourner a l'accueil
	ctrlH := &desktop.CustomShortcut{
		KeyName:  fyne.KeyH,
		Modifier: fyne.KeyModifierControl,
	}
	a.fenetre.Canvas().AddShortcut(ctrlH, func(shortcut fyne.Shortcut) {
		a.afficherAccueil()
	})
}

// afficherAccueil - affiche la page d'accueil avec la grille d'artistes
func (a *AppGroupie) afficherAccueil() {
	page := a.creerPageAccueil()
	a.fenetre.SetContent(page)
}

// afficherDetail - affiche la page de detail d'un artiste
func (a *AppGroupie) afficherDetail(artiste models.Artiste) {
	page := a.creerPageDetail(artiste)
	a.fenetre.SetContent(page)
}

// getImageArtiste - recupere l'image d'un artiste depuis le cache ou l'API
func (a *AppGroupie) getImageArtiste(artiste models.Artiste) []byte {
	a.cacheImagesMu.RLock()
	if data, ok := a.cacheImages[artiste.ID]; ok {
		a.cacheImagesMu.RUnlock()
		return data
	}
	a.cacheImagesMu.RUnlock()

	data, err := api.RecupererImageArtiste(artiste.Image)
	if err != nil {
		fmt.Println("Erreur image pour", artiste.Nom, ":", err)
		return nil
	}

	a.cacheImagesMu.Lock()
	a.cacheImages[artiste.ID] = data
	a.cacheImagesMu.Unlock()

	return data
}

// toggleFavori - ajoute ou enleve un artiste des favoris
func (a *AppGroupie) toggleFavori(id int) {
	a.favorisMu.Lock()
	if a.favoris[id] {
		delete(a.favoris, id)
	} else {
		a.favoris[id] = true
	}
	a.favorisMu.Unlock()
}

// estFavori - verifie si un artiste est dans les favoris
func (a *AppGroupie) estFavori(id int) bool {
	a.favorisMu.RLock()
	defer a.favorisMu.RUnlock()
	return a.favoris[id]
}

// creerBoutonRetour - cree un bouton retour vers la page d'accueil
func (a *AppGroupie) creerBoutonRetour() *widget.Button {
	btn := widget.NewButtonWithIcon("Retour", theme.NavigateBackIcon(), func() {
		a.afficherAccueil()
	})
	return btn
}

// creerHeader - cree le header commun avec le titre et les raccourcis
func creerHeader() *fyne.Container {
	titre := widget.NewLabel("🎵 Groupie Tracker")
	titre.TextStyle = fyne.TextStyle{Bold: true}

	raccourcisInfo := widget.NewLabel("Ctrl+F: Recherche | Ctrl+H: Accueil")
	raccourcisInfo.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewHBox(titre, layout.NewSpacer(), raccourcisInfo)
}
