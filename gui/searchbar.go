package gui

import (
	"fmt"
	"strings"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// searchbar.go - la barre de recherche avec suggestions en temps reel
// c'est un peu le truc le plus chiant a faire mais ca marche bien au final

// EntryRecherche - un widget Entry customise pour la recherche
// avec les suggestions qui s'affichent en dessous
type EntryRecherche struct {
	widget.Entry
	onChanged func(string)
}

// NewEntryRecherche - cree une nouvelle barre de recherche
func NewEntryRecherche() *EntryRecherche {
	e := &EntryRecherche{}
	e.ExtendBaseWidget(e)
	e.PlaceHolder = "🔍 Rechercher un artiste, membre, lieu..."
	return e
}

// genererSuggestions - genere toutes les suggestions basees sur le texte tape
// gere les cas: nom d'artiste, membres, locations, date 1er album, date creation
func genererSuggestions(texte string, artistes []models.Artiste, locData models.IndexLocations) []models.SuggestionRecherche {
	if texte == "" {
		return nil
	}

	var suggestions []models.SuggestionRecherche
	texteMin := strings.ToLower(strings.TrimSpace(texte))

	// on va checker chaque artiste pour chaque type de recherche
	for _, artiste := range artistes {
		// recherche par nom d'artiste/groupe
		if strings.Contains(strings.ToLower(artiste.Nom), texteMin) {
			suggestions = append(suggestions, models.SuggestionRecherche{
				Texte:     artiste.Nom + " → artist/band",
				Type:      "artist/band",
				ArtisteID: artiste.ID,
			})
		}

		// recherche par membre
		for _, membre := range artiste.Membres {
			membreClean := strings.TrimSpace(membre)
			if strings.Contains(strings.ToLower(membreClean), texteMin) {
				suggestions = append(suggestions, models.SuggestionRecherche{
					Texte:     membreClean + " → member (" + artiste.Nom + ")",
					Type:      "member",
					ArtisteID: artiste.ID,
				})
			}
		}

		// recherche par date du premier album
		if strings.Contains(strings.ToLower(artiste.PremierAlbum), texteMin) {
			suggestions = append(suggestions, models.SuggestionRecherche{
				Texte:     artiste.PremierAlbum + " → first album date (" + artiste.Nom + ")",
				Type:      "first album date",
				ArtisteID: artiste.ID,
			})
		}

		// recherche par date de creation
		dateCreStr := fmt.Sprintf("%d", artiste.DateCreation)
		if strings.Contains(dateCreStr, texteMin) {
			suggestions = append(suggestions, models.SuggestionRecherche{
				Texte:     fmt.Sprintf("%d → creation date (%s)", artiste.DateCreation, artiste.Nom),
				Type:      "creation date",
				ArtisteID: artiste.ID,
			})
		}
	}

	// recherche par lieu de concert
	dejavu := make(map[string]bool)
	for _, loc := range locData.Index {
		for _, lieu := range loc.Locations {
			lieuPropre := strings.ReplaceAll(lieu, "_", " ")
			lieuPropre = strings.ReplaceAll(lieuPropre, "-", ", ")
			lieuMin := strings.ToLower(lieuPropre)

			cleUnique := fmt.Sprintf("%s_%d", lieuPropre, loc.ID)
			if strings.Contains(lieuMin, texteMin) && !dejavu[cleUnique] {
				dejavu[cleUnique] = true

				// trouver l'artiste correspondant
				nomArtiste := "?"
				for _, a := range artistes {
					if a.ID == loc.ID {
						nomArtiste = a.Nom
						break
					}
				}
				suggestions = append(suggestions, models.SuggestionRecherche{
					Texte:     lieuPropre + " → location (" + nomArtiste + ")",
					Type:      "location",
					ArtisteID: loc.ID,
				})
			}
		}
	}

	// on limite a 15 suggestions pas plus sinon c'est le bordel
	if len(suggestions) > 15 {
		suggestions = suggestions[:15]
	}

	return suggestions
}

// creerPanneauSuggestions - cree le panneau d'affichage des suggestions
func creerPanneauSuggestions(suggestions []models.SuggestionRecherche, onSelect func(models.SuggestionRecherche)) *fyne.Container {
	items := make([]fyne.CanvasObject, 0, len(suggestions))

	for _, s := range suggestions {
		suggestion := s // capture pour la closure, sinon bug classique
		btn := widget.NewButton(suggestion.Texte, func() {
			onSelect(suggestion)
		})
		btn.Importance = widget.LowImportance
		btn.Alignment = widget.ButtonAlignLeading
		items = append(items, btn)
	}

	return container.NewVBox(items...)
}
