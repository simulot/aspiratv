# aspiratv

Ce programme interroge les serveurs de télévision de rattrapage et télécharge les émissions souhaitées selon une organisation reconnue par le programme [PLEX](https://www.plex.tv/).

## Avertissement
Aspiratv ne fait que garder une copie de l'oeuvre sur votre disque dur, comme vous l'auriez fait avec votre enregistreur video, votre box TV ou une clé USB branchée sur votre TV. Cette opération est seulement rendue plus simple qu'en gérant manuellement les enregistrements.

## Prérequis

- FFMPEG: ffmpeg est utilisé pour convertir le flux video en fichiers mp4. l'exécutable doit être diponible dans votre système. Page de téléchargement pour Windows: [https://ffmpeg.zeranoe.com/builds/](https://ffmpeg.zeranoe.com/builds/)


## Configuration

### fichier **config.json**

Le fichier config.json contient les paramètres et la liste des émissions que l'on shouhaite télécharger :

``` json
{
  "PullInterval": "1h30m",
  "Destinations": {
    "Documentaires": "${HOME}/Videos/Documentaires",
    "Jeunesse": "${HOME}/Videos/Jeunesse"
  },
  "WatchList": [
    {
      "Show": "Garfield",
      "Title": "",
      "Pitch": "",
      "Provider": "francetv",
      "Destination": "Jeunesse"
    },
    {
      "Show": "Les routes de l'impossible",
      "Title": "",
      "Pitch": "",
      "Provider": "francetv",
      "Destination": "Documentaires"
    }    
  ]
}
```
### PullInterval
Intervalle entre deux recherches sur le serveur de la télévision, selon le format "1h30" pour un intervalle d'une heure trente.
Le délai exact est aléatoire pour ne pas interroger le serveur à heures fixes.

### Destinations
Défini les répertoires de destination des fichiers. A noter que les variables d'environnement peuvent être utilisées.

### WatchList
Donne la liste des critères de recherche pour sélectionner les émissions à télécharger. L'ensemble des critères non vides doit être statisfait. Ils sont évalués dans l'ordre suivant :
1. Provider: code du fournisseur de contenu
1. Show : nom de l'émission
1. Title: titre de l'émission ou de l'épisode
1. Pitch: description de l'émission
Le contenu du critère doit être contenu dans le champ correspondant obtenu sur le serveur de la télévision.

* Destination: code du répertoire où les fichiers doivent être téléchargés, dont la définition est placée dans la section  **Destinations**


## Les fournisseurs de contenu: les providers
Un provider est un package du logiciel permettant d'implémenter les différents connecteurs.
Les connecteurs disponibles sont :
* France Télévision (`francetv`):
  * Programmes en replay des chaine France 2, France 3, France 4, France 5, France Ô, La 1ère
* Arte France (`artetv`) :
  * Programmes en langue française ou sous-titrés en français.

## Configuration de PLEX

Pour obtenir un résultat acceptable, il faut configurer une librairie de type "Séries TV" en utilisant l'agent "Personal Media Shows" afin que plex utilise les titres et les imagettes téléchargées depuis le server de la télévision. Veillez à ce que l'agent "Local Media Assets (TV)" soit placé en tête de liste des agents pour les Séries / Personal Media Shows ([voir cette page](https://support.plex.tv/articles/200265256-naming-home-series-media/)) . 

 

# Téléchargement

Les binaires pour Windows, Linux et FreeBSD sont directement disponibles sur la page [releases](https://github.com/simulot/aspiratv/releases/latest)

# Compilation
Vous devez avoir un compilateur pour [le langage GO](https://golang.org/dl/).


