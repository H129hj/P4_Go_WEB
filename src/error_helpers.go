package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type errorPageData struct {
	Code    string
	Message string
}

// Redirection vers une page d'erreur customisée //
func redirectToError(w http.ResponseWriter, r *http.Request, code int, message string) {
	params := url.Values{}
	if code > 0 {
		params.Set("code", strconv.Itoa(code))
	}
	if strings.TrimSpace(message) != "" {
		params.Set("message", message)
	}

	target := "/error"
	if encoded := params.Encode(); encoded != "" {
		target += "?" + encoded
	}

	http.Redirect(w, r, target, http.StatusSeeOther)
}

func handleErrorDisplay(w http.ResponseWriter, r *http.Request) {
	data := errorPageData{
		Code:    r.FormValue("code"),
		Message: r.FormValue("message"),
	}
	renderTemplate(w, r, "Error", data)
}
