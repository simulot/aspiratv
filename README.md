# aspiratv

Ce programme interroge les serveurs de télévision de rattrapage et télécharge les émissions souhaitées selon une organisation reconnue par le programme [PLEX](https://www.plex.tv/).

## Avertissement
Les contenus mis à disposition par les diffuseurs sont soumis aux droits d'auteur. Ne les utilisez pas en dehors du cadre privé.

Aspiratv ne fait que garder une copie de l’œuvre sur votre disque dur, comme vous l'auriez fait avec votre enregistreur vidéo, votre box TV ou une clé USB branchée sur votre TV. Cette opération est seulement rendue plus simple qu'en gérant manuellement les enregistrements.

Le fonctionnement de ce programme n'est pas garanti. Notamment, les fournisseurs de contenus sont susceptibles de changer leurs APIs ou interdire leur utilisation sans pré-avis. 

## Prérequis

- FFMPEG: ffmpeg est utilisé pour convertir le flux vidéo en fichiers mp4. l'exécutable doit être disponible dans votre système. Page de téléchargement pour Windows: [https://ffmpeg.zeranoe.com/builds/](https://ffmpeg.zeranoe.com/builds/)

# Installation

Les binaires pour Windows, Linux et FreeBSD sont directement disponibles sur la page [releases](https://github.com/simulot/aspiratv/releases/latest). Les binaires n'ont pas de dépendance autre que FFMPEG et n'ont pas besoin d'être installés.

# Ligne de commande

```
Usage of ./aspiratv:
  -debug
        Debug mode.
  -force
        Force media download.
  -service
        Run as service.
```
## -debug
Augmente le nombre de message de log.

## -force
Télécharge toutes les émissions correspondant à la liste de recherche, même si elles ont été déjà téléchargées.

## -service
Dans ce mode, le programme reste actif et interroge les serveurs régulièrement.


# Configuration

## fichier **config.json**

Le fichier config.json contient les paramètres et la liste des émissions que l'on souhaite télécharger :

``` json
{
  "PullInterval": "1h30m",
  "Destinations": {
    "Documentaires": "${HOME}/Videos/Documentaires",
    "Jeunesse": "${HOME}/Videos/Jeunesse",
    "Courts": "${HOME}/Videos/Courts"
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
    },
    {
      "Playlist": "La minute vieille",
      "Title": "",
      "Pitch": "",
      "Provider": "artetv",
      "Destination": "Courts"
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
Donne la liste des critères de recherche pour sélectionner les émissions à télécharger. L'ensemble des critères non vides doit être satisfait. Ils sont évalués dans l'ordre suivant :
1. Provider: code du fournisseur de contenu
1. Show : nom de l'émission
1. Title: titre de l'émission ou de l'épisode
1. Pitch: description de l'émission
Le contenu du critère doit être contenu dans le champ correspondant obtenu sur le serveur de la télévision.

* Destination: code du répertoire où les fichiers doivent être téléchargés, dont la définition est placée dans la section  **Destinations**

Chaque provider peut traiter spécifiquement les recherches. 

# Les fournisseurs de contenu : les providers
Un provider est un package du logiciel permettant d'implémenter les différents connecteurs.
Les connecteurs disponibles sont :
* France Télévision (`francetv`):
  * Programmes en replay des chaîne France 2, France 3, France 4, France 5, France Ô, La 1ère
* Arte France (`artetv`) :
  * Programmes en langue française ou sous-titrés en français.
  * Les playlists Arte peuvent être surveillées pour que les nouveaux épisodes soit téléchargés dès leur disponibilité. 

# Configuration de PLEX

Pour obtenir un résultat acceptable, il faut configurer une librairie de type "Séries TV" en utilisant l'agent "Personal Media Shows" afin que plex utilise les titres et les imagettes téléchargées depuis le serveur de la télévision. Veillez à ce que l'agent "Local Media Assets (TV)" soit placé en tête de liste des agents pour les Séries / Personal Media Shows ([voir cette page](https://support.plex.tv/articles/200265256-naming-home-series-media/)) . 


# Compilation des sources
Vous devez avoir un compilateur pour [le langage GO](https://golang.org/dl/).

# Todo

- [x] Provider pour Arte
- [x] Arte.TV: Suivre les collections
- [_] Provider pour Gulli 

