package gui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// filters.go - les filtres pour la page d'accueil
// on a des sliders pour les dates et des checkboxes pour les membres / locations

// Filtres - contient les valeurs actuelles des filtres
type Filtres struct {
	CreationMin int
	CreationMax int
	AlbumMin    int
	AlbumMax    int
	NbMembres   map[int]bool    // les nombres de membres coches
	Locations   map[string]bool // les locations cochees
}

// NewFiltres - cree des filtres par defaut (tout est ouvert)
func NewFiltres() *Filtres {
	return &Filtres{
		CreationMin: 1950,
		CreationMax: 2025,
		AlbumMin:    1950,
		AlbumMax:    2025,
		NbMembres:   make(map[int]bool),
		Locations:   make(map[string]bool),
	}
}

// extraireAnneePremierAlbum - parse l'annee depuis le format "DD-MM-YYYY" du premier album
func extraireAnneePremierAlbum(dateStr string) int {
	parts := strings.Split(dateStr, "-")
	if len(parts) == 3 {
		annee, err := strconv.Atoi(parts[2])
		if err == nil {
			return annee
		}
	}
	return 0
}

// appliquerFiltres - filtre les artistes selon les criteres choisis
func appliquerFiltres(artistes []models.Artiste, filtres *Filtres, locData models.IndexLocations) []models.Artiste {
	var resultat []models.Artiste

	for _, artiste := range artistes {
		// filtre par date de creation
		if artiste.DateCreation < filtres.CreationMin || artiste.DateCreation > filtres.CreationMax {
			continue
		}

		// filtre par annee du premier album
		anneeAlbum := extraireAnneePremierAlbum(artiste.PremierAlbum)
		if anneeAlbum > 0 && (anneeAlbum < filtres.AlbumMin || anneeAlbum > filtres.AlbumMax) {
			continue
		}

		// filtre par nombre de membres (si des checkboxes sont cochees)
		if len(filtres.NbMembres) > 0 {
			nbMembres := len(artiste.Membres)
			// on check si le nb de membres correspond, 7 = "7 ou plus"
			trouve := false
			for nb := range filtres.NbMembres {
				if nb == 7 && nbMembres >= 7 {
					trouve = true
					break
				}
				if nb == nbMembres {
					trouve = true
					break
				}
			}
			if !trouve {
				continue
			}
		}

		// filtre par location (si des locations sont cochees)
		if len(filtres.Locations) > 0 {
			trouveLoc := false
			for _, loc := range locData.Index {
				if loc.ID == artiste.ID {
					for _, lieu := range loc.Locations {
						lieuPropre := strings.ReplaceAll(lieu, "_", " ")
						lieuPropre = strings.ReplaceAll(lieuPropre, "-", ", ")
						// on check si la location correspond
						for locFiltre := range filtres.Locations {
							if strings.Contains(strings.ToLower(lieuPropre), strings.ToLower(locFiltre)) {
								trouveLoc = true
								break
							}
						}
						if trouveLoc {
							break
						}
					}
					break
				}
			}
			if !trouveLoc {
				continue
			}
		}

		resultat = append(resultat, artiste)
	}

	return resultat
}

// extraireTousLesPays - extrait tous les pays uniques depuis les locations
func extraireTousLesPays(locData models.IndexLocations) []string {
	paysMap := make(map[string]bool)
	for _, loc := range locData.Index {
		for _, lieu := range loc.Locations {
			parts := strings.Split(lieu, "-")
			if len(parts) >= 2 {
				pays := parts[len(parts)-1]
				pays = strings.ReplaceAll(pays, "_", " ")
				paysMap[pays] = true
			}
		}
	}

	var pays []string
	for p := range paysMap {
		pays = append(pays, p)
	}
	sort.Strings(pays)
	return pays
}

// creerPanneauFiltres - cree le panneau lateral avec tous les filtres
func creerPanneauFiltres(filtres *Filtres, locData models.IndexLocations, onFiltreChange func()) fyne.CanvasObject {
	// === FILTRE DATE DE CREATION (range slider) ===
	labelCreation := widget.NewLabel(fmt.Sprintf("Création: %d - %d", filtres.CreationMin, filtres.CreationMax))
	labelCreation.TextStyle = fyne.TextStyle{Bold: true}

	sliderCreationMin := widget.NewSlider(1950, 2025)
	sliderCreationMin.Value = float64(filtres.CreationMin)
	sliderCreationMin.Step = 1

	sliderCreationMax := widget.NewSlider(1950, 2025)
	sliderCreationMax.Value = float64(filtres.CreationMax)
	sliderCreationMax.Step = 1

	sliderCreationMin.OnChanged = func(val float64) {
		filtres.CreationMin = int(val)
		if filtres.CreationMin > filtres.CreationMax {
			filtres.CreationMax = filtres.CreationMin
			sliderCreationMax.Value = val
			sliderCreationMax.Refresh()
		}
		labelCreation.SetText(fmt.Sprintf("Création: %d - %d", filtres.CreationMin, filtres.CreationMax))
		onFiltreChange()
	}
	sliderCreationMax.OnChanged = func(val float64) {
		filtres.CreationMax = int(val)
		if filtres.CreationMax < filtres.CreationMin {
			filtres.CreationMin = filtres.CreationMax
			sliderCreationMin.Value = val
			sliderCreationMin.Refresh()
		}
		labelCreation.SetText(fmt.Sprintf("Création: %d - %d", filtres.CreationMin, filtres.CreationMax))
		onFiltreChange()
	}

	filtreCreation := container.NewVBox(
		labelCreation,
		widget.NewLabel("Min:"),
		sliderCreationMin,
		widget.NewLabel("Max:"),
		sliderCreationMax,
	)

	// === FILTRE PREMIER ALBUM (range slider) ===
	labelAlbum := widget.NewLabel(fmt.Sprintf("1er Album: %d - %d", filtres.AlbumMin, filtres.AlbumMax))
	labelAlbum.TextStyle = fyne.TextStyle{Bold: true}

	sliderAlbumMin := widget.NewSlider(1950, 2025)
	sliderAlbumMin.Value = float64(filtres.AlbumMin)
	sliderAlbumMin.Step = 1

	sliderAlbumMax := widget.NewSlider(1950, 2025)
	sliderAlbumMax.Value = float64(filtres.AlbumMax)
	sliderAlbumMax.Step = 1

	sliderAlbumMin.OnChanged = func(val float64) {
		filtres.AlbumMin = int(val)
		if filtres.AlbumMin > filtres.AlbumMax {
			filtres.AlbumMax = filtres.AlbumMin
			sliderAlbumMax.Value = val
			sliderAlbumMax.Refresh()
		}
		labelAlbum.SetText(fmt.Sprintf("1er Album: %d - %d", filtres.AlbumMin, filtres.AlbumMax))
		onFiltreChange()
	}
	sliderAlbumMax.OnChanged = func(val float64) {
		filtres.AlbumMax = int(val)
		if filtres.AlbumMax < filtres.AlbumMin {
			filtres.AlbumMin = filtres.AlbumMax
			sliderAlbumMin.Value = val
			sliderAlbumMin.Refresh()
		}
		labelAlbum.SetText(fmt.Sprintf("1er Album: %d - %d", filtres.AlbumMin, filtres.AlbumMax))
		onFiltreChange()
	}

	filtreAlbum := container.NewVBox(
		labelAlbum,
		widget.NewLabel("Min:"),
		sliderAlbumMin,
		widget.NewLabel("Max:"),
		sliderAlbumMax,
	)

	// === FILTRE NOMBRE DE MEMBRES (checkboxes) ===
	labelMembres := widget.NewLabel("Nombre de membres:")
	labelMembres.TextStyle = fyne.TextStyle{Bold: true}

	membresChecks := container.NewVBox(labelMembres)
	for i := 1; i <= 7; i++ {
		nb := i
		label := fmt.Sprintf("%d", nb)
		if nb == 7 {
			label = "7+"
		}
		check := widget.NewCheck(label, func(checked bool) {
			if checked {
				filtres.NbMembres[nb] = true
			} else {
				delete(filtres.NbMembres, nb)
			}
			onFiltreChange()
		})
		membresChecks.Add(check)
	}

	// === FILTRE LOCATIONS (checkboxes des pays) ===
	labelLocations := widget.NewLabel("Pays:")
	labelLocations.TextStyle = fyne.TextStyle{Bold: true}

	paysListe := extraireTousLesPays(locData)
	locChecks := container.NewVBox(labelLocations)
	for _, pays := range paysListe {
		p := pays // capture
		check := widget.NewCheck(p, func(checked bool) {
			if checked {
				filtres.Locations[p] = true
			} else {
				delete(filtres.Locations, p)
			}
			onFiltreChange()
		})
		locChecks.Add(check)
	}

	// bouton reset pour tout remettre a zero
	btnReset := widget.NewButton("🔄 Reset filtres", func() {
		filtres.CreationMin = 1950
		filtres.CreationMax = 2025
		filtres.AlbumMin = 1950
		filtres.AlbumMax = 2025
		filtres.NbMembres = make(map[int]bool)
		filtres.Locations = make(map[string]bool)
		onFiltreChange()
	})
	btnReset.Importance = widget.HighImportance

	// on met tout dans un container scrollable
	contenuFiltres := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel("⚙️ FILTRES"),
		widget.NewSeparator(),
		filtreCreation,
		widget.NewSeparator(),
		filtreAlbum,
		widget.NewSeparator(),
		membresChecks,
		widget.NewSeparator(),
		locChecks,
		widget.NewSeparator(),
		btnReset,
		layout.NewSpacer(),
	)

	scrollFiltres := container.NewVScroll(contenuFiltres)
	scrollFiltres.SetMinSize(fyne.NewSize(250, 400))

	return scrollFiltres
}
