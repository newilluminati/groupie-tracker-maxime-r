package geo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"groupie-tracker/models"
)

// geocode.go - la geolocalisation des lieux de concert
// on utilise Nominatim (OpenStreetMap) qui est gratuit et sans cle API
// c'est pas le plus rapide mais ca fait le taf

// le client HTTP pour les requetes de geocoding (timeout un peu plus long)
var geoClient = &http.Client{
	Timeout: 10 * time.Second,
}

// cache pour eviter de refaire les memes requetes encore et encore
// on stocke les coordonnees des villes deja geocodees
var (
	cacheCoords = make(map[string]models.Coordonnees)
	cacheMutex  sync.RWMutex
)

// reponseNominatim - la structure de la reponse de l'API Nominatim
type reponseNominatim struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// NettoyerLieu - prend un lieu de l'API genre "north_carolina-usa" et le transforme
// en quelque chose de plus propre genre "north carolina, usa" pour la recherche
func NettoyerLieu(lieu string) string {
	// on remplace les underscores par des espaces et les tirets par des virgules
	lieu = strings.ReplaceAll(lieu, "_", " ")
	// le dernier tiret separe la ville/region du pays
	idx := strings.LastIndex(lieu, "-")
	if idx != -1 {
		lieu = lieu[:idx] + ", " + lieu[idx+1:]
	}
	return lieu
}

// Geocoder - convertit une adresse en coordonnees GPS
// utilise le cache si on a deja cherche cette adresse
func Geocoder(adresse string) (models.Coordonnees, error) {
	// on regarde d'abord dans le cache
	cacheMutex.RLock()
	if coords, ok := cacheCoords[adresse]; ok {
		cacheMutex.RUnlock()
		return coords, nil
	}
	cacheMutex.RUnlock()

	// pas dans le cache, on fait la requete a Nominatim
	// faut respecter leur rate limit (1 requete par seconde)
	reqURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?format=json&q=%s&limit=1",
		url.QueryEscape(adresse),
	)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return models.Coordonnees{}, fmt.Errorf("erreur creation requete: %w", err)
	}
	// Nominatim demande un User-Agent valide sinon il bloque
	req.Header.Set("User-Agent", "GroupieTracker-Student-Project/1.0")

	resp, err := geoClient.Do(req)
	if err != nil {
		return models.Coordonnees{}, fmt.Errorf("erreur requete geocoding pour '%s': %w", adresse, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Coordonnees{}, fmt.Errorf("erreur lecture reponse geocoding: %w", err)
	}

	var resultats []reponseNominatim
	err = json.Unmarshal(body, &resultats)
	if err != nil {
		return models.Coordonnees{}, fmt.Errorf("erreur parsing reponse geocoding: %w", err)
	}

	if len(resultats) == 0 {
		return models.Coordonnees{}, fmt.Errorf("aucun resultat pour l'adresse '%s'", adresse)
	}

	// on parse les coordonnees
	var lat, lng float64
	fmt.Sscanf(resultats[0].Lat, "%f", &lat)
	fmt.Sscanf(resultats[0].Lon, "%f", &lng)

	coords := models.Coordonnees{Lat: lat, Lng: lng}

	// on met dans le cache pour la prochaine fois
	cacheMutex.Lock()
	cacheCoords[adresse] = coords
	cacheMutex.Unlock()

	return coords, nil
}

// GeocoderLieuAPI - prend un lieu de l'API et le geocode
// c'est un raccourci qui nettoie le lieu avant de le geocoder
func GeocoderLieuAPI(lieuAPI string) (models.Coordonnees, error) {
	adressePropre := NettoyerLieu(lieuAPI)
	return Geocoder(adressePropre)
}
