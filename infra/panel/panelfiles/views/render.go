package views

import (
	"net/http"
	"time"

	"github.com/CSPF-Founder/iva/panel/config"
	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/internal/sessions"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/utils"
	"github.com/CSPF-Founder/iva/panel/views/helpers"
	"github.com/CSPF-Founder/iva/panel/views/layouts"
	"github.com/a-h/templ"
)

func NewBaseData(conf *config.Config, session *sessions.SessionManager, r *http.Request) helpers.BaseData {
	checkUser := ctx.Get(r, "user")
	user := models.User{}
	if checkUser != nil {
		user = ctx.Get(r, "user").(models.User)
	}
	year, _, _ := time.Now().Date()

	return helpers.BaseData{
		CSRFName:               conf.ServerConf.CSRFName,
		CSRFToken:              session.GetCSRF(r.Context()),
		User:                   user,
		Version:                config.Version,
		Flashes:                session.Flashes(r.Context()),
		ProductTitle:           conf.ProductTitle,
		CopyrightFooterCompany: conf.CopyrightFooterCompany,
		CurrentYear:            year,
		PreviousPage:           utils.GetRelativePath(r),
	}
}

func RenderTempl(
	component templ.Component,
	data helpers.BaseData,
	w http.ResponseWriter,
	r *http.Request,
) error {
	return layouts.BaseLayout(component, data).Render(r.Context(), w)
}

func RenderPlainTempl(
	component templ.Component,
	data helpers.BaseData,
	w http.ResponseWriter,
	r *http.Request,
) error {
	return layouts.PlainBodyLayout(component, data).Render(r.Context(), w)
}
