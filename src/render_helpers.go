package main

import (
	"bytes"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	var buffer bytes.Buffer
	if err := temp.ExecuteTemplate(&buffer, name, data); err != nil {
		redirectToError(w, r, http.StatusInternalServerError, "Erreur lors du chargement de la page")
		return
	}

	_, _ = buffer.WriteTo(w)
}
