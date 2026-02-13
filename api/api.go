package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"groupie-tracker/models"
)

// api.go - tout ce qui est fetch de données depuis l'API groupie tracker
// on utilise un client HTTP avec timeout pour pas bloquer l'app si l'API est lente

// l'url de base de l'API, on la met en constante pour pas la retaper partout
const baseURL = "https://groupietrackers.herokuapp.com/api"

// le client HTTP avec un timeout de 15 sec (on est pas pressés mais on veut pas attendre 3h non plus)
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// fetchJSON - fonction generique pour fetch du JSON depuis une URL
// elle gere les erreurs HTTP et le parsing JSON, c'est pratique
func fetchJSON(url string, cible interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("erreur requete HTTP vers %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a repondu %d pour %s, c'est pas normal", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erreur lecture du body: %w", err)
	}

	err = json.Unmarshal(body, cible)
	if err != nil {
		return fmt.Errorf("erreur parsing JSON depuis %s: %w", url, err)
	}

	return nil
}

// RecupererArtistes - va chercher tous les artistes depuis l'API
// retourne un slice d'Artiste et une erreur si ca foire
func RecupererArtistes() ([]models.Artiste, error) {
	var artistes []models.Artiste
	err := fetchJSON(baseURL+"/artists", &artistes)
	if err != nil {
		return nil, fmt.Errorf("impossible de recuperer les artistes: %w", err)
	}
	return artistes, nil
}

// RecupererRelation - va chercher la relation pour un artiste particulier
// c'est la qu'on a les concerts avec dates + lieux
func RecupererRelation(id int) (models.Relation, error) {
	var relation models.Relation
	url := fmt.Sprintf("%s/relation/%d", baseURL, id)
	err := fetchJSON(url, &relation)
	if err != nil {
		return relation, fmt.Errorf("impossible de recuperer la relation pour l'artiste %d: %w", id, err)
	}
	return relation, nil
}

// RecupererToutesLocations - va chercher toutes les locations de tous les artistes
// utile pour les filtres par lieu de concert
func RecupererToutesLocations() (models.IndexLocations, error) {
	var locs models.IndexLocations
	err := fetchJSON(baseURL+"/locations", &locs)
	if err != nil {
		return locs, fmt.Errorf("impossible de recuperer les locations: %w", err)
	}
	return locs, nil
}

// RecupererImageArtiste - telecharge l'image d'un artiste et retourne les bytes
// on fait ca pour afficher les images dans Fyne
func RecupererImageArtiste(imageURL string) ([]byte, error) {
	resp, err := httpClient.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("erreur telechargement image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("l'image a repondu %d, bizarre", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture image: %w", err)
	}

	return data, nil
}
