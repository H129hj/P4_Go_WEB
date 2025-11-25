package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Gamestate : l'état de la partie en cours
type GameState struct {
	Grid          [6][7]string
	PlayerNames   [2]string
	PlayerTokens  [2]string
	CurrentPlayer int
	Winner        string
	Draw          bool
	Initialized   bool
}

// GamePage : Donnees à passer au template de la page de jeu
type GamePage struct {
	Grille            [6][7]string
	Joueur1           string
	Joueur2           string
	JetonCouleur      string
	CurrentPlayerName string
	CurrentToken      string
	Winner            string
	Draw              bool
	Message           string
	Columns           []int
}

var tmpl *template.Template
var gameState GameState

// Démarre le serveur et configure les routes
func main() {
	var err error
	tmpl, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	rootDoc, _ := os.Getwd()
	fileServer := http.FileServer(http.Dir(rootDoc + "/assets"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", handleHomepage)
	http.HandleFunc("/game/init", handleInitPage)
	http.HandleFunc("/game/init/traitement", handleInitSubmit)
	http.HandleFunc("/game/play", handleGamePlay)
	http.HandleFunc("/game/play/move", handleMove)
	http.HandleFunc("/game/scoreboards", handleScoreboards)

	http.ListenAndServe("localhost:8000", nil)
}

// Handlers pour les différentes routes
func handleHomepage(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "Homepage", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleInitPage(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "GameInit", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleInitSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupération et validation des données du formulaire
	j1 := strings.TrimSpace(r.FormValue("name"))
	j2 := strings.TrimSpace(r.FormValue("name2"))
	jetonCouleur := r.FormValue("jetoncolor")

	if j1 == "" || j2 == "" {
		http.Redirect(w, r, "/game/init", http.StatusSeeOther)
		return
	}

	if jetonCouleur != "rouge" && jetonCouleur != "jaune" {
		jetonCouleur = "rouge"
	}

	gameState.Reset(j1, j2, jetonCouleur)

	http.Redirect(w, r, "/game/play", http.StatusSeeOther)
}

func handleGamePlay(w http.ResponseWriter, r *http.Request) {
	// Vérification que le jeu est initialisé
	if !gameState.Initialized {
		http.Redirect(w, r, "/game/init", http.StatusSeeOther)
		return
	}

	// Données pour le template de la page de jeu
	page := buildPageData(gameState, r.URL.Query().Get("msg"))

	if err := tmpl.ExecuteTemplate(w, "GamePlay", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


// Handler pour le traitement des coups joués
func handleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	col, err := strconv.Atoi(r.FormValue("column"))
	if err != nil {
		http.Redirect(w, r, "/game/play?msg="+url.QueryEscape("Choisissez une colonne valide."), http.StatusSeeOther)
		return
	}

	if dropErr := gameState.Drop(col); dropErr != nil {
		http.Redirect(w, r, "/game/play?msg="+url.QueryEscape(dropErr.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/game/play", http.StatusSeeOther)
}

func handleScoreboards(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "Scoreboards", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Réinitialise l'état du jeu avec la page game/init
func (g *GameState) Reset(j1, j2, jeton string) {
	g.Grid = [6][7]string{}
	g.PlayerNames = [2]string{j1, j2}
	secondColor := "rouge"
	if jeton == "rouge" {
		secondColor = "jaune"
	}
	g.PlayerTokens = [2]string{jeton, secondColor}
	g.CurrentPlayer = 0
	g.Winner = ""
	g.Draw = false
	g.Initialized = true
}

// Logique pour déposer un jeton dans une colonne et vérifie l'état du jeu (gagnant, match nul)
func (g *GameState) Drop(column int) error {
	if g.Winner != "" || g.Draw {
		return errors.New("La partie est terminée.")
	}

	if column < 0 || column >= len(g.Grid[0]) {
		return errors.New("Colonne inexistante.")
	}

	for row := len(g.Grid) - 1; row >= 0; row-- {
		if g.Grid[row][column] == "" {
			token := g.PlayerTokens[g.CurrentPlayer]
			g.Grid[row][column] = token

			if g.hasWinner(row, column, token) {
				g.Winner = g.PlayerNames[g.CurrentPlayer]
			} else if g.isBoardFull() {
				g.Draw = true
			} else {
				g.CurrentPlayer = 1 - g.CurrentPlayer
			}
			return nil
		}
	}

	return errors.New("Cette colonne est pleine.")
}

// Vérifie si un joueur a gagné
func (g *GameState) hasWinner(row, col int, token string) bool {
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions {
		total := 1 + g.countDirection(row, col, dir[0], dir[1], token) + g.countDirection(row, col, -dir[0], -dir[1], token)
		if total >= 4 {
			return true
		}
	}

	return false
}


// Compte le nombre de jetons consécutifs dans une direction donnée
func (g *GameState) countDirection(row, col, dr, dc int, token string) int {
	count := 0
	rCur := row + dr
	cCur := col + dc

	for rCur >= 0 && rCur < len(g.Grid) && cCur >= 0 && cCur < len(g.Grid[0]) && g.Grid[rCur][cCur] == token {
		count++
		rCur += dr
		cCur += dc
	}

	return count
}

// Vérifie si le plateau est plein (match nul)
func (g *GameState) isBoardFull() bool {
	for _, cell := range g.Grid[0] {
		if cell == "" {
			return false
		}
	}

	return true
}

// Construit les données à passer au template de la page de jeu
func buildPageData(state GameState, message string) GamePage {
	columns := make([]int, len(state.Grid[0]))
	for i := range columns {
		columns[i] = i
	}

	currentName := state.PlayerNames[state.CurrentPlayer]
	currentToken := state.PlayerTokens[state.CurrentPlayer]
	if state.Winner != "" || state.Draw {
		currentName = ""
		currentToken = ""
	}

	return GamePage{
		Grille:            state.Grid,
		Joueur1:           state.PlayerNames[0],
		Joueur2:           state.PlayerNames[1],
		JetonCouleur:      state.PlayerTokens[0],
		CurrentPlayerName: currentName,
		CurrentToken:      currentToken,
		Winner:            state.Winner,
		Draw:              state.Draw,
		Message:           message,
		Columns:           columns,
	}
}
