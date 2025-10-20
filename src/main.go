package main

import (
	"fmt"
	"net/http"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	http.HandleFunc("/Homepage", func(w http.ResponseWriter, r *http.Request) {
		err := listTemplate.ExecuteTemplate(w, "Homepage.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe("localhost:8000", nil)
}