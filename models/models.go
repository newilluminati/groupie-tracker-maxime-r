package models

// models.go - les structures de données pour l'API groupie tracker
// on met tout ici pour pas se prendre la tete a chercher partout

// Artiste - c'est la struct principale, elle contient toutes les infos d'un artiste/groupe
type Artiste struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Nom          string   `json:"name"`
	Membres      []string `json:"members"`
	DateCreation int      `json:"creationDate"`
	PremierAlbum string   `json:"firstAlbum"`
	LocationsURL string   `json:"locations"`
	DatesURL     string   `json:"concertDates"`
	RelationsURL string   `json:"relations"`
}

// Relation - le lien entre les dates et les lieux de concert
// genre c'est la qu'on sait quel concert etait ou et quand
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// IndexLocations - la reponse de l'API /locations qui contient tous les lieux
type IndexLocations struct {
	Index []LocationData `json:"index"`
}

// LocationData - les lieux de concert d'un artiste
type LocationData struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"`
}

// Coordonnees - pour stocker la latitude et longitude apres le geocoding
type Coordonnees struct {
	Lat float64
	Lng float64
}

// SuggestionRecherche - une suggestion dans la barre de recherche
// avec le texte a afficher et le type (member, artist/band, location, etc)
type SuggestionRecherche struct {
	Texte     string // le texte a afficher genre "Phil Collins"
	Type      string // le type: "artist/band", "member", "location", "first album date", "creation date"
	ArtisteID int    // l'id de l'artiste correspondant pour pouvoir naviguer
}
