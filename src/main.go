package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

func main() {
	type GamePage struct {
		Grille [6][7]string
		Tour   int
		Joueur1 string
		Joueur2 string
		JetonCouleur string
	}

	type GameData struct {
		Joueur1     string
		Joueur2     string
		JetonCouleur string
	}

	var currentGame GameData

	listTemplate, errTemp := template.ParseGlob("./templates/*.html")
	if errTemp != nil {
		fmt.Println(errTemp.Error())
		os.Exit(1)
	}

	rootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(rootDoc + "/assets"))
	http.Handle("/static/", http.StripPrefix("/static/", fileserver))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := listTemplate.ExecuteTemplate(w, "Homepage", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/game/init", func(w http.ResponseWriter, r *http.Request) {
		err := listTemplate.ExecuteTemplate(w, "GameInit", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/game/init/traitement", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    j1 := r.FormValue("name")
    j2 := r.FormValue("name2")
    jetonCouleur := r.FormValue("jetoncolor")

    currentGame = GameData{
        Joueur1:      j1,
        Joueur2:      j2,
        JetonCouleur: jetonCouleur,
    }

    http.Redirect(w, r, "/game/play", http.StatusSeeOther)
})


	http.HandleFunc("/game/play", func(w http.ResponseWriter, r *http.Request) {
		data := GamePage{
			Grille: [6][7]string{{"", "", "", "", "", "", ""}, {"", "", "", "", "", "", ""}, {"", "", "", "", "", "", ""}, {"", "", "J", "R", "R", "", "J"}, {"", "R", "J", "R", "R", "", "R"}, {"J", "R", "J", "J", "J", "R", "J"}},
			Tour:   1,
			Joueur1: currentGame.Joueur1,
			Joueur2: currentGame.Joueur2,
			JetonCouleur: currentGame.JetonCouleur,
		}

		listTemplate.ExecuteTemplate(w, "GamePlay", data)
	})

	http.ListenAndServe("localhost:8000", nil)
}
