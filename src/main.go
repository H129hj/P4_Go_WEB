package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type GameState struct {
	Grid          [6][7]string
	PlayerNames   [2]string
	PlayerTokens  [2]string
	CurrentPlayer int
	Winner        string
	Draw          bool
	Initialized   bool
	TurnCount     int
}

type GamePage struct {
	Grille             [6][7]string
	Joueur1            string
	Joueur2            string
	JetonCouleur       string
	CurrentPlayerName  string
	CurrentPlayerIndex int
	CurrentToken       string
	Winner             string
	Draw               bool
	Message            string
	Columns            []int
}

type EndPage struct {
	Grille       [6][7]string
	Joueur1      string
	Joueur2      string
	JetonCouleur string
	Winner       string
	Draw         bool
}

type GameRecord struct {
	ID           int          `json:"id"`
	Joueur1      string       `json:"joueur1"`
	Joueur2      string       `json:"joueur2"`
	Winner       string       `json:"winner"`
	Draw         bool         `json:"draw"`
	Date         time.Time    `json:"date"`
	TurnCount    int          `json:"turnCount"`
	Grille       [6][7]string `json:"grille"`
	JetonCouleur string       `json:"jetonCouleur"`
}

type LeaderboardPage struct {
	Records []GameRecord `json:"records"`
}

type GameGridPage struct {
	Record GameRecord
}

var tmpl *template.Template
var gameState GameState

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
	http.HandleFunc("/game/end", handleGameEnd)
	http.HandleFunc("/game/leaderboard", handleLeaderboard)
	http.HandleFunc("/game/grid/", handleGameGrid)
	http.HandleFunc("/error", handleErrorDisplay)

	http.ListenAndServe("localhost:8000", nil)
}

func handleHomepage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "Homepage", nil)
}

func handleInitPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "GameInit", nil)
}

func handleInitSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		redirectToError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	j1 := strings.TrimSpace(r.FormValue("name"))
	j2 := strings.TrimSpace(r.FormValue("name2"))
	jetonCouleur := r.FormValue("jetoncolor")

	if j1 == "" || j2 == "" {
		redirectToError(w, r, http.StatusBadRequest, "Les noms des joueurs sont requis")
		return
	}

	if jetonCouleur != "rouge" && jetonCouleur != "jaune" {
		jetonCouleur = "rouge"
	}

	gameState.Reset(j1, j2, jetonCouleur)

	http.Redirect(w, r, "/game/play", http.StatusSeeOther)
}

func handleGamePlay(w http.ResponseWriter, r *http.Request) {

	if !gameState.Initialized {
		redirectToError(w, r, http.StatusBadRequest, "La partie n'a pas été initialisée")
		return
	}

	page := buildPageData(gameState, r.URL.Query().Get("msg"))

	renderTemplate(w, r, "GamePlay", page)
}

func handleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		redirectToError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	col, err := strconv.Atoi(strings.TrimSpace(r.FormValue("column")))
	if err != nil {
		redirectToError(w, r, http.StatusBadRequest, "Choisissez une colonne valide.")
		return
	}

	if !gameState.Initialized {
		redirectToError(w, r, http.StatusBadRequest, "La partie n'a pas été initialisée")
		return
	}

	if dropErr := gameState.Drop(col); dropErr != nil {
		http.Redirect(w, r, "/game/play?msg="+url.QueryEscape(dropErr.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/game/play", http.StatusSeeOther)
}

func handleGameEnd(w http.ResponseWriter, r *http.Request) {
	if !gameState.Initialized {
		redirectToError(w, r, http.StatusBadRequest, "La partie n'a pas été initialisée")
		return
	}

	if gameState.Winner == "" && !gameState.Draw {
		redirectToError(w, r, http.StatusBadRequest, "La partie n'est pas terminée")
		return
	}

	// Sauvegarder la partie dans le leaderboard
	saveGameRecord()

	page := EndPage{
		Grille:       gameState.Grid,
		Joueur1:      gameState.PlayerNames[0],
		Joueur2:      gameState.PlayerNames[1],
		JetonCouleur: gameState.PlayerTokens[0],
		Winner:       gameState.Winner,
		Draw:         gameState.Draw,
	}

	renderTemplate(w, r, "GameEnd", page)
}

func saveGameRecord() {
	// Lire les records existants pour obtenir le prochain ID
	records := loadGameRecords()
	nextID := 1
	if len(records) > 0 {
		maxID := 0
		for _, r := range records {
			if r.ID > maxID {
				maxID = r.ID
			}
		}
		nextID = maxID + 1
	}

	record := GameRecord{
		ID:           nextID,
		Joueur1:      gameState.PlayerNames[0],
		Joueur2:      gameState.PlayerNames[1],
		Winner:       gameState.Winner,
		Draw:         gameState.Draw,
		Date:         time.Now(),
		TurnCount:    gameState.TurnCount,
		Grille:       gameState.Grid,
		JetonCouleur: gameState.PlayerTokens[0],
	}

	records = append(records, record)

	// Sauvegarder dans le fichier
	file, err := os.OpenFile("leaderboard.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du fichier: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(records); err != nil {
		fmt.Printf("Erreur lors de l'écriture: %v\n", err)
		return
	}
}

func loadGameRecords() []GameRecord {
	file, err := os.Open("leaderboard.txt")
	if err != nil {
		// Si le fichier n'existe pas, retourner une liste vide
		return []GameRecord{}
	}
	defer file.Close()

	var records []GameRecord
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&records); err != nil {
		// Si erreur de décodage, retourner une liste vide
		return []GameRecord{}
	}

	// Assigner des IDs aux anciennes parties qui n'en ont pas
	for i := range records {
		if records[i].ID == 0 {
			records[i].ID = i + 1
		}
	}

	return records
}

func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	records := loadGameRecords()

	// Inverser l'ordre pour afficher les plus récents en premier
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	page := LeaderboardPage{
		Records: records,
	}

	renderTemplate(w, r, "Leaderboard", page)
}

func handleGameGrid(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID de l'URL (format: /game/grid/123)
	path := strings.TrimPrefix(r.URL.Path, "/game/grid/")
	id, err := strconv.Atoi(path)
	if err != nil {
		redirectToError(w, r, http.StatusBadRequest, "ID invalide")
		return
	}

	records := loadGameRecords()
	var foundRecord *GameRecord
	for _, record := range records {
		if record.ID == id {
			foundRecord = &record
			break
		}
	}

	if foundRecord == nil {
		redirectToError(w, r, http.StatusNotFound, "Partie non trouvée")
		return
	}

	page := GameGridPage{
		Record: *foundRecord,
	}

	renderTemplate(w, r, "GameGrid", page)
}

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
	g.TurnCount = 0
}

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
			g.TurnCount++

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

func (g *GameState) isBoardFull() bool {
	for _, cell := range g.Grid[0] {
		if cell == "" {
			return false
		}
	}

	return true
}

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

	currentPlayerIndex := -1
	if state.Winner == "" && !state.Draw {
		currentPlayerIndex = state.CurrentPlayer
	}

	return GamePage{
		Grille:             state.Grid,
		Joueur1:            state.PlayerNames[0],
		Joueur2:            state.PlayerNames[1],
		JetonCouleur:       state.PlayerTokens[0],
		CurrentPlayerName:  currentName,
		CurrentPlayerIndex: currentPlayerIndex,
		CurrentToken:       currentToken,
		Winner:             state.Winner,
		Draw:               state.Draw,
		Message:            message,
		Columns:            columns,
	}
}
