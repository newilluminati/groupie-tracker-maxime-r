## Groupie Tracker

Un projet en Go ou on recupere des donnees d'artistes/groupes depuis une API et on les affiche dans une appli avec Fyne.

## Comment lancer

il faut avoir Go d'installe puis:

go run .

## Ce que l'app fait

- elle affiche tous les artistes dans une grille avec leur photo et leur nom
- y'a une barre de recherche qui propose des suggestions quand on tape (artiste, membre, lieu, date etc)
- on peut filtrer par date de creation, premier album, nombre de membres ou par pays
- quand on clique sur un artiste ca ouvre sa page avec ses infos et ses concerts
- les concerts sont affiches sur une carte grace a la geolocalisation (on utilise Nominatim)
- on peut mettre des artistes en favoris
- y'a des raccourcis clavier (Ctrl+F pour chercher, Ctrl+H pour revenir a l'accueil)

## Comment c'est organise

- main.go -> c'est le fichier principal qui lance l'app
- api/api.go -> c'est la ou on va chercher les donnees sur l'API
- models/models.go -> les structures pour stocker les donnees des artistes
- gui/ -> tout ce qui est interface (la page d'accueil, la page detail, la recherche, les filtres)
- geo/geocode.go -> la geolocalisation des concerts

## Technologies

- Go
- Fyne pour l'interface graphique
- Nominatim (OpenStreetMap) pour la geolocalisation
