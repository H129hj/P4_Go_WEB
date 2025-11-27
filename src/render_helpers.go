package main

import (
	"bytes"
	"net/http"
)

// renderTemplate centralises template execution with buffering to avoid
// partially written responses and redirects to /error when rendering fails.
func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	var buffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buffer, name, data); err != nil {
		redirectToError(w, r, http.StatusInternalServerError, "Erreur lors du chargement de la page")
		return
	}

	_, _ = buffer.WriteTo(w)
}
