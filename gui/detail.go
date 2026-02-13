package gui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"groupie-tracker/api"
	"groupie-tracker/geo"
	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// detail.go - la page de detail d'un artiste
// quand tu cliques sur un artiste, tu arrives ici avec toutes ses infos

// PointCarte - un point a afficher sur la carte avec son nom et ses coordonnees
type PointCarte struct {
	Lieu   string
	Coords models.Coordonnees
}

// creerPageDetail - construit la page complete de detail d'un artiste
func (a *AppGroupie) creerPageDetail(artiste models.Artiste) fyne.CanvasObject {
	// bouton retour
	btnRetour := a.creerBoutonRetour()

	// header avec bouton retour et titre
	titreDetail := widget.NewLabel("🎤 " + artiste.Nom)
	titreDetail.TextStyle = fyne.TextStyle{Bold: true}

	// bouton favori
	etoileTxt := "☆ Ajouter aux favoris"
	if a.estFavori(artiste.ID) {
		etoileTxt = "⭐ Retirer des favoris"
	}
	btnFav := widget.NewButton(etoileTxt, func() {
		a.toggleFavori(artiste.ID)
		// on rafraichit la page pour mettre a jour le bouton
		a.afficherDetail(artiste)
	})

	headerDetail := container.NewHBox(btnRetour, titreDetail, layout.NewSpacer(), btnFav)

	// === SECTION IMAGE ===
	var imgArtiste *canvas.Image
	imageData := a.getImageArtiste(artiste)
	if imageData != nil {
		imgRes := fyne.NewStaticResource(fmt.Sprintf("detail_%d", artiste.ID), imageData)
		imgArtiste = canvas.NewImageFromResource(imgRes)
		imgArtiste.FillMode = canvas.ImageFillContain
		imgArtiste.SetMinSize(fyne.NewSize(300, 300))
	}

	// === SECTION INFOS ===
	labelCreation := widget.NewLabel(fmt.Sprintf("📅 Année de création: %d", artiste.DateCreation))
	labelAlbum := widget.NewLabel(fmt.Sprintf("💿 Premier album: %s", artiste.PremierAlbum))
	labelNbMembres := widget.NewLabel(fmt.Sprintf("👥 Nombre de membres: %d", len(artiste.Membres)))

	// liste des membres
	labelMembresTitle := widget.NewLabel("🎸 Membres:")
	labelMembresTitle.TextStyle = fyne.TextStyle{Bold: true}
	membresContainer := container.NewVBox(labelMembresTitle)
	for _, membre := range artiste.Membres {
		m := strings.TrimSpace(membre)
		membresContainer.Add(widget.NewLabel("  • " + m))
	}

	infosContainer := container.NewVBox(
		labelCreation,
		labelAlbum,
		labelNbMembres,
		widget.NewSeparator(),
		membresContainer,
	)

	// la partie haute: image a gauche, infos a droite
	var partieHaute fyne.CanvasObject
	if imgArtiste != nil {
		partieHaute = container.NewHBox(
			imgArtiste,
			widget.NewSeparator(),
			infosContainer,
		)
	} else {
		partieHaute = infosContainer
	}

	// === SECTION CONCERTS (on fetch la relation) ===
	concertsContainer := container.NewVBox(
		widget.NewLabel("⏳ Chargement des concerts..."),
	)

	// === SECTION CARTE ===
	carteContainer := container.NewVBox(
		widget.NewLabel("🗺️ Chargement de la carte..."),
	)

	// on fetch les donnees de relation en arriere-plan
	go func() {
		relation, err := api.RecupererRelation(artiste.ID)
		if err != nil {
			concertsContainer.RemoveAll()
			concertsContainer.Add(widget.NewLabel(fmt.Sprintf("❌ Erreur: %v", err)))
			concertsContainer.Refresh()
			return
		}

		// afficher les concerts
		concertsContainer.RemoveAll()
		labelConcerts := widget.NewLabel("🎵 Concerts:")
		labelConcerts.TextStyle = fyne.TextStyle{Bold: true}
		concertsContainer.Add(labelConcerts)

		if len(relation.DatesLocations) == 0 {
			concertsContainer.Add(widget.NewLabel("  Aucun concert trouvé"))
		} else {
			for lieu, dates := range relation.DatesLocations {
				lieuPropre := formaterLieu(lieu)
				for _, date := range dates {
					labelConcert := widget.NewLabel(fmt.Sprintf("  📍 %s  —  📅 %s", lieuPropre, date))
					concertsContainer.Add(labelConcert)
				}
			}
		}
		concertsContainer.Refresh()

		// maintenant on fait la carte avec les geocoords
		a.chargerCarte(relation, carteContainer)
	}()

	// assembler la page complete
	contenuDetail := container.NewVBox(
		headerDetail,
		widget.NewSeparator(),
		partieHaute,
		widget.NewSeparator(),
		concertsContainer,
		widget.NewSeparator(),
		carteContainer,
		layout.NewSpacer(),
	)

	// on met tout dans un scroll
	scroll := container.NewVScroll(contenuDetail)
	return scroll
}

// formaterLieu - formate un lieu de l'API en quelque chose de lisible
// "north_carolina-usa" -> "North Carolina, Usa"
func formaterLieu(lieu string) string {
	lieu = strings.ReplaceAll(lieu, "_", " ")
	idx := strings.LastIndex(lieu, "-")
	if idx != -1 {
		lieu = lieu[:idx] + ", " + lieu[idx+1:]
	}
	// mettre chaque mot en majuscule (fait main car strings.Title est deprecated)
	mots := strings.Fields(lieu)
	for i, mot := range mots {
		if len(mot) > 0 {
			mots[i] = strings.ToUpper(mot[:1]) + mot[1:]
		}
	}
	return strings.Join(mots, " ")
}

// chargerCarte - geocode les lieux et dessine la carte
func (a *AppGroupie) chargerCarte(relation models.Relation, carteContainer *fyne.Container) {
	var points []PointCarte

	for lieu := range relation.DatesLocations {
		lieuPropre := geo.NettoyerLieu(lieu)
		coords, err := geo.GeocoderLieuAPI(lieu)
		if err != nil {
			fmt.Printf("Geocoding echoue pour '%s': %v\n", lieuPropre, err)
			continue
		}
		points = append(points, PointCarte{Lieu: lieuPropre, Coords: coords})
		// on attend un peu entre chaque requete pour respecter le rate limit de Nominatim
		time.Sleep(1100 * time.Millisecond)
	}

	// afficher la carte
	carteContainer.RemoveAll()

	if len(points) == 0 {
		carteContainer.Add(widget.NewLabel("🗺️ Aucun lieu géolocalisé"))
		carteContainer.Refresh()
		return
	}

	labelCarte := widget.NewLabel("🗺️ Carte des concerts:")
	labelCarte.TextStyle = fyne.TextStyle{Bold: true}
	carteContainer.Add(labelCarte)

	// on dessine une carte simple avec les points
	carteWidget := dessinerCarte(points)
	carteContainer.Add(carteWidget)

	// on ajoute la legende en dessous
	for _, pt := range points {
		label := widget.NewLabel(fmt.Sprintf("  📍 %s (%.2f, %.2f)", pt.Lieu, pt.Coords.Lat, pt.Coords.Lng))
		carteContainer.Add(label)
	}

	carteContainer.Refresh()
}

// dessinerCarte - dessine une carte du monde simple avec des points rouges
// c'est un canvas custom qui fait une projection plate des coordonnees
func dessinerCarte(points []PointCarte) fyne.CanvasObject {
	largeur := float32(700)
	hauteur := float32(400)

	// le fond de la carte (rectangle bleu fonce pour l'ocean)
	fond := canvas.NewRectangle(color.RGBA{R: 20, G: 30, B: 50, A: 255})
	fond.SetMinSize(fyne.NewSize(largeur, hauteur))

	// container pour les elements de la carte
	elements := []fyne.CanvasObject{fond}

	// des rectangles verts pour representer les continents (c'est simplifie)
	continentsData := []struct {
		x, y, w, h float32
	}{
		{0.10, 0.15, 0.15, 0.20}, // amerique du nord
		{0.18, 0.40, 0.08, 0.25}, // amerique du sud
		{0.45, 0.15, 0.12, 0.15}, // europe
		{0.47, 0.35, 0.10, 0.25}, // afrique
		{0.55, 0.10, 0.25, 0.25}, // asie
		{0.78, 0.55, 0.08, 0.10}, // australie
	}

	for _, c := range continentsData {
		rect := canvas.NewRectangle(color.RGBA{R: 60, G: 100, B: 60, A: 150})
		rect.Resize(fyne.NewSize(largeur*c.w, hauteur*c.h))
		rect.Move(fyne.NewPos(largeur*c.x, hauteur*c.y))
		elements = append(elements, rect)
	}

	// on dessine l'equateur et le meridien pour orienter
	equateur := canvas.NewLine(color.RGBA{R: 100, G: 100, B: 100, A: 100})
	equateur.Position1 = fyne.NewPos(0, hauteur/2)
	equateur.Position2 = fyne.NewPos(largeur, hauteur/2)
	elements = append(elements, equateur)

	meridien := canvas.NewLine(color.RGBA{R: 100, G: 100, B: 100, A: 100})
	meridien.Position1 = fyne.NewPos(largeur/2, 0)
	meridien.Position2 = fyne.NewPos(largeur/2, hauteur)
	elements = append(elements, meridien)

	// ajouter les points de concert sur la carte
	for _, pt := range points {
		// conversion lat/lng en position x/y (projection plate simple)
		x := float32((pt.Coords.Lng + 180) / 360 * float64(largeur))
		y := float32((90 - pt.Coords.Lat) / 180 * float64(hauteur))

		// clamp pour rester dans les bornes
		if x < 0 {
			x = 0
		}
		if x > largeur {
			x = largeur
		}
		if y < 0 {
			y = 0
		}
		if y > hauteur {
			y = hauteur
		}

		// halo lumineux autour du point
		halo := canvas.NewCircle(color.RGBA{R: 255, G: 100, B: 100, A: 80})
		halo.Resize(fyne.NewSize(20, 20))
		halo.Move(fyne.NewPos(x-10, y-10))
		elements = append(elements, halo)

		// le point rouge du concert
		point := canvas.NewCircle(color.RGBA{R: 255, G: 50, B: 50, A: 255})
		point.Resize(fyne.NewSize(10, 10))
		point.Move(fyne.NewPos(x-5, y-5))
		elements = append(elements, point)
	}

	carte := container.NewWithoutLayout(elements...)
	carte.Resize(fyne.NewSize(largeur, hauteur))

	return container.NewStack(carte)
}
