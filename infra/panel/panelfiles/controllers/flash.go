package controllers

import (
	"net/http"

	"github.com/CSPF-Founder/iva/panel/internal/sessions"
	"github.com/CSPF-Founder/iva/panel/utils"
)

// flash handles the rendering flash messages
func (app *App) flash(_ http.ResponseWriter, r *http.Request, t string, m string, c bool) {
	app.session.AddFlash(r.Context(), sessions.SessionFlash{
		Type:     t,
		Message:  m,
		Closable: c,
	})
}

func (app *App) flashAndGoBack(w http.ResponseWriter, r *http.Request, t string, message string) {
	app.flash(w, r, t, message, true)
	utils.RedirectBack(w, r)
}
