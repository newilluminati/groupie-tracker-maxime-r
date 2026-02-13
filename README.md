# Groupie Tracker

Application desktop en Go qui affiche les infos d'artistes/groupes depuis l'API Groupie Trackers.

## Lancer

```
go run .
```

## Features

- Grille d'artistes avec images, nom, annee de creation
- Barre de recherche avec suggestions (artiste, membre, lieu, date)
- Filtres par date de creation, premier album, nombre de membres, pays
- Page detail avec infos completes + liste des concerts
- Carte des concerts avec geolocalisation (Nominatim/OpenStreetMap)
- Systeme de favoris
- Raccourcis clavier (Ctrl+F recherche, Ctrl+H accueil)

## Stack

- Go
- Fyne (GUI)
- API Nominatim pour la geolocalisation

## Structure

```
main.go          point d'entree
api/api.go       fetch des donnees API
models/models.go structures de donnees
gui/             interface graphique (accueil, detail, recherche, filtres)
geo/geocode.go   geolocalisation
```
