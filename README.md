# P4_Web_Hugo_Shakil
Ce projet est une implémentation web du jeu Puissance 4 en Go. Il sert de support pour pratiquer Go, le templating HTML et la persistance simple côté serveur.

## Comment lancer le serveur :
```bash
git clone https://github.com/H129hj/P4_Go_WEB
cd P4_Go_WEB
```

```bash
cd src
go run .
```

## Prèrequis 
Installation du language GO 

## Description
Puissance 4 Web reprend les règles classiques (grille 6x7, deux joueurs). L’application propose un parcours complet : accueil, configuration des joueurs, partie, fin et historique des matchs.

## Fonctionnalités
- Interface web responsive (HTML/CSS)
- Saisie des pseudos + choix visuel de la couleur du joueur 1
- Indicateur de tour et boutons colonne alignés à la grille
- Redirection automatique vers la page de fin (timer 1,5 s)
- Leaderboard persistant (`leaderboard.txt`) avec date, vainqueur, nombre de tours
- Accès à la grille finale depuis le leaderboard
- Détection des victoires horizontales, verticales, diagonales + détection du nul

## Comment jouer
1. Ouvrez `http://localhost:8000`
2. Cliquez sur **Commencer une partie**
3. Renseignez les pseudos, choisissez la couleur du joueur 1 puis validez
4. Depuis la page de jeu, cliquez sur une colonne pour lâcher un pion
5. Attendez la fin de partie ou la redirection automatique pour consulter le résultat
6. Utilisez le leaderboard pour revoir les parties passées et leurs grilles

## Commandes
- Boutons fléchés sous chaque colonne : déposer un jeton
- **Nouvelle partie** : retour à la configuration
- **Leaderboard** : liste les parties et propose “Voir la grille”

## Structure du projet
- `assets/css/` : styles globaux (`style.css`, `game.css`)
- `templates/` : vues HTML (`Homepage`, `GameInit`, `GamePlay`, `GameEnd`, `Leaderboard`, `GameGrid`)
- `leaderboard.txt` : stockage JSON des parties terminées
- `src/main.go` : serveur HTTP, logique métier, persistance

## Notes techniques
- Serveur Go standard (`net/http`)
- Template réalisé en html + css
- Grille stockée en mémoire puis sérialisée en JSON
- Leaderboard géré via `encoding/json` sur un fichier local

## Développé par
- BERTON Hugo
- KHALDI Shakil
