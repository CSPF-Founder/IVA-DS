package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/CSPF-Founder/iva/panel/auth"
	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/enums"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/utils"
	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/pages/userpages"
	"github.com/go-chi/chi/v5"
)

type userController struct {
	*App
}

func newUserController(a *App) *userController {
	return &userController{a}
}

func (c *userController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Login Routes
	router.Get("/login", mid.Use(c.displayLogin))
	router.Post("/login", mid.Use(c.loginHandler))

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		r.Use(mid.RequireLogin)
		r.Get("/profile", c.profile)
		r.Get("/logout", c.logout)
	})

	return router
}

// logout destroys the current user session
func (c *userController) logout(w http.ResponseWriter, r *http.Request) {
	c.session.Remove(r.Context(), "userID")
	if err := c.session.Destroy(r.Context()); err != nil {
		c.logger.Error("Error destroying session ", err)
	}
	c.flash(w, r, "success", "You have successfully logged out", true)

	http.Redirect(w, r, "/users/login", http.StatusFound)
}

// Login handles the authentication flow for a user. If credentials are valid,
// a session is created
func (c *userController) displayLogin(w http.ResponseWriter, r *http.Request) {
	hasAnyUser, _ := models.HasAnyUsers()
	if !hasAnyUser {
		http.Redirect(w, r, "/setup/create-user", http.StatusSeeOther)
		return
	}

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Login"
	if err := views.RenderPlainTempl(userpages.Login(templateData), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template", err)
	}

}

// Redirect to login page if login attempt is invalid
func (c *userController) handleInvalidLogin(w http.ResponseWriter, r *http.Request, message string) {
	c.flash(w, r, enums.FlashWarning, message, true)
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func doLogin(r *http.Request) (models.User, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	u, err := models.GetUserByUsername(username)
	if err != nil {
		return u, fmt.Errorf("Invalid Username/Password")
	}
	// Validate the user's password
	err = auth.ValidatePassword(password, u.Password)
	if err != nil {
		return u, fmt.Errorf("Invalid Username/Password")
	}
	return u, nil
}

func (c *userController) loginHandler(w http.ResponseWriter, r *http.Request) {
	u, err := doLogin(r)
	if err != nil {
		c.logger.Error("Login Error", err)
		c.handleInvalidLogin(w, r, "Invalid Username/Password")
		return
	}

	// First renew the session token to prevent session fixation
	if err := c.session.RenewToken(r.Context()); err != nil {
		c.logger.Error("Error Renew Session Token at login ", err)
	}

	c.session.Put(r.Context(), "userID", u.ID)

	c.nextOrIndex(w, r)

}

func (c *userController) nextOrIndex(w http.ResponseWriter, r *http.Request) {
	next := "/"
	url, err := url.Parse(r.FormValue("next"))
	if err == nil {
		path := url.EscapedPath()
		if path != "" {
			next = "/" + strings.TrimLeft(path, "/")
		}
	}
	if !utils.IsRelativeURL(next) {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func (c *userController) profile(w http.ResponseWriter, r *http.Request) {

	user := ctx.Get(r, "user").(models.User)

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Profile"
	if err := views.RenderTempl(userpages.Profile(user), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template", err)
	}
}
