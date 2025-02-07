package controllers

import (
	"net/http"

	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/errpages"
	"github.com/a-h/templ"
)

type errorController struct {
	*App
}

func newErrorController(a *App) *errorController {
	return &errorController{a}
}

func (c *errorController) forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	isAjaxRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	if isAjaxRequest {
		c.SendJSONError(w, "Access Denied", http.StatusForbidden)
		return
	}
	c.flash(w, r, "danger", "You do not have permission to access this page.", true)

	w.WriteHeader(http.StatusForbidden)

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Access Denied"

	if err := views.RenderPlainTempl(errpages.ForbiddenPage(templateData.Flashes, templ.SafeURL(templateData.PreviousPage)), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

func (c *errorController) notFoundHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNotFound)

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "404 Not Found"

	if err := views.RenderPlainTempl(errpages.NotFound(templateData.Flashes), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

func (c *errorController) csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
	isAjaxRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	if isAjaxRequest {
		c.SendJSONError(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	c.flash(w, r, "danger", "Invalid CSRF token", true)

	w.WriteHeader(http.StatusForbidden)

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Access Denied"

	if err := views.RenderPlainTempl(errpages.ForbiddenPage(templateData.Flashes, templ.SafeURL(templateData.PreviousPage)), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}
