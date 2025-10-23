package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

func main() {
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
		http.Redirect(w, r, "game/play", http.StatusSeeOther)
	})

	http.HandleFunc("/game/play", func(w http.ResponseWriter, r *http.Request) {
		err := listTemplate.ExecuteTemplate(w, "Gameplay", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe("localhost:8000", nil)
}
