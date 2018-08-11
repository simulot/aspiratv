# aspiratv

Ce programme interroge les serveurs de télévision de rattrapage et télécharge les émissions souhaitées.

## Configuration

Le fichier config.json contient les paramètres et la liste des émissions souhaitées:

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
Intervalle entre deux recherches sur le serveur de la télévision

### Destinations
Défini les répertoires de destination des fichiers. A noter que les variables d'environnement peuvent être ête utiliser.

### WatchList
Donne la liste des critères de recherche pour sélectionner les émissions à télécharger. L'ensemble des critères non vide doit être statisfait:
- Provider: code du fournisseur de contenu
- Show : nom de l'émission
- Title: titre de l'émission ou de l'épisode
- Pitch: déscription de l'émission
Le contenu du critère doit être contenu dans le champ correspondant obtenu sur le serveur de la télévision.

* Destination: code du répertoire où les fichiers doivent être téléchargés, définis dans **Destinations**





