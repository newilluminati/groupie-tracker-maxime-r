package gui

import (
	"fmt"
	"strings"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// home.go - la page d'accueil avec la grille de cards des artistes
// c'est la premiere page qu'on voit quand on lance l'app

// creerPageAccueil - construit toute la page d'accueil
// avec la recherche en haut, les filtres a gauche et la grille au centre
func (a *AppGroupie) creerPageAccueil() fyne.CanvasObject {
	// le header avec le titre
	header := creerHeader()

	// les filtres
	filtres := NewFiltres()

	// la grille des artistes (on la cree d'abord vide)
	grille := container.NewGridWrap(fyne.NewSize(220, 380))

	// container pour les suggestions
	containerSuggestions := container.NewVBox()

	// la barre de recherche
	entryRecherche := NewEntryRecherche()
	a.barreRecherche = entryRecherche

	// variable pour le texte de recherche actuel
	texteRecherche := ""

	// fonction pour rafraichir la grille avec les filtres et la recherche
	rafraichirGrille := func() {
		// on applique d'abord les filtres
		artistesFiltres := appliquerFiltres(a.artistes, filtres, a.locationsData)

		// puis la recherche textuelle si y'a un texte
		if texteRecherche != "" {
			var artistesRecherche []models.Artiste
			texteMin := strings.ToLower(texteRecherche)
			for _, art := range artistesFiltres {
				// on check dans le nom
				if strings.Contains(strings.ToLower(art.Nom), texteMin) {
					artistesRecherche = append(artistesRecherche, art)
					continue
				}
				// on check dans les membres
				trouveMembre := false
				for _, m := range art.Membres {
					if strings.Contains(strings.ToLower(m), texteMin) {
						trouveMembre = true
						break
					}
				}
				if trouveMembre {
					artistesRecherche = append(artistesRecherche, art)
					continue
				}
				// on check la date de creation
				if strings.Contains(fmt.Sprintf("%d", art.DateCreation), texteMin) {
					artistesRecherche = append(artistesRecherche, art)
					continue
				}
				// on check le premier album
				if strings.Contains(strings.ToLower(art.PremierAlbum), texteMin) {
					artistesRecherche = append(artistesRecherche, art)
				}
			}
			artistesFiltres = artistesRecherche
		}

		// on reconstruit la grille
		grille.RemoveAll()
		for _, artiste := range artistesFiltres {
			art := artiste // capture pour la closure
			card := a.creerCardArtiste(art)
			grille.Add(card)
		}
		grille.Refresh()
	}

	// callback quand on tape dans la recherche
	entryRecherche.OnChanged = func(texte string) {
		texteRecherche = texte

		// generer les suggestions
		containerSuggestions.RemoveAll()
		if texte != "" {
			suggestions := genererSuggestions(texte, a.artistes, a.locationsData)
			panneau := creerPanneauSuggestions(suggestions, func(s models.SuggestionRecherche) {
				// quand on clique sur une suggestion, on va sur l'artiste
				for _, art := range a.artistes {
					if art.ID == s.ArtisteID {
						a.afficherDetail(art)
						return
					}
				}
			})
			containerSuggestions.Add(panneau)
		}
		containerSuggestions.Refresh()

		rafraichirGrille()
	}

	// stocker le callback de rafraichissement
	a.onRefreshAccueil = rafraichirGrille

	// callback quand les filtres changent
	onFiltreChange := func() {
		rafraichirGrille()
	}

	// construire le panneau de filtres
	panneauFiltres := creerPanneauFiltres(filtres, a.locationsData, onFiltreChange)

	// afficher la grille initiale
	rafraichirGrille()

	// label nb de resultats
	labelResultats := widget.NewLabel(fmt.Sprintf("%d artistes", len(a.artistes)))

	// layout de la recherche
	barreRechercheContainer := container.NewVBox(
		container.NewBorder(nil, nil, nil, nil, entryRecherche),
		containerSuggestions,
	)

	// la zone principale avec la grille scrollable
	scrollGrille := container.NewVScroll(grille)
	scrollGrille.SetMinSize(fyne.NewSize(600, 500))

	// assemble tout: header en haut, filtres a gauche, grille au centre
	zoneHaut := container.NewVBox(
		header,
		widget.NewSeparator(),
		barreRechercheContainer,
		labelResultats,
		widget.NewSeparator(),
	)

	contenu := container.NewBorder(
		zoneHaut,       // haut
		nil,            // bas
		panneauFiltres, // gauche
		nil,            // droite
		scrollGrille,   // centre
	)

	return contenu
}

// creerCardArtiste - cree une card pour un artiste dans la grille
// avec son image, son nom et l'annee de creation
func (a *AppGroupie) creerCardArtiste(artiste models.Artiste) fyne.CanvasObject {
	// on charge l'image en arriere plan
	var imgWidget *canvas.Image

	imageData := a.getImageArtiste(artiste)
	if imageData != nil {
		imgRes := fyne.NewStaticResource(fmt.Sprintf("artist_%d", artiste.ID), imageData)
		imgWidget = canvas.NewImageFromResource(imgRes)
		imgWidget.FillMode = canvas.ImageFillContain
		imgWidget.SetMinSize(fyne.NewSize(150, 150))
	} else {
		// image par defaut si on arrive pas a charger
		imgWidget = canvas.NewImageFromResource(nil)
		imgWidget.SetMinSize(fyne.NewSize(150, 150))
	}

	// le nom de l'artiste
	labelNom := widget.NewLabel(artiste.Nom)
	labelNom.TextStyle = fyne.TextStyle{Bold: true}
	labelNom.Alignment = fyne.TextAlignCenter
	labelNom.Wrapping = fyne.TextWrapWord

	// l'annee de creation
	labelAnnee := widget.NewLabel(fmt.Sprintf("📅 %d", artiste.DateCreation))
	labelAnnee.Alignment = fyne.TextAlignCenter

	// etoile favori
	etoileTxt := "☆"
	if a.estFavori(artiste.ID) {
		etoileTxt = "⭐"
	}
	btnFavori := widget.NewButton(etoileTxt, func() {
		a.toggleFavori(artiste.ID)
		// on rafraichit l'accueil pour mettre a jour les etoiles
		if a.onRefreshAccueil != nil {
			a.afficherAccueil()
		}
	})
	btnFavori.Importance = widget.LowImportance

	// bouton pour voir le detail
	btnDetail := widget.NewButton("Voir détails", func() {
		a.afficherDetail(artiste)
	})
	btnDetail.Importance = widget.MediumImportance

	// on assemble le tout dans un container vertical
	cardContent := container.NewVBox(
		imgWidget,
		labelNom,
		labelAnnee,
		container.NewHBox(layout.NewSpacer(), btnFavori, layout.NewSpacer()),
		btnDetail,
	)

	return widget.NewCard("", "", cardContent)
}
