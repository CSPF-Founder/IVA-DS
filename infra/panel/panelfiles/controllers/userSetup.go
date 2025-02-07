package controllers

import (
	"net/http"
	"strings"

	"github.com/CSPF-Founder/iva/panel/auth"
	"github.com/CSPF-Founder/iva/panel/enums"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/utils"
	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/pages/userpages"
	"github.com/go-chi/chi/v5"
)

type userSetupController struct {
	*App
}

func newUserSetupController(a *App) *userSetupController {
	return &userSetupController{a}
}

func (c *userSetupController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(c.userSetupMiddleware)
		r.Get("/create-user", mid.Use(c.displayCreateUser))
		r.Post("/create-user", mid.Use(c.createUserHandler))
	})

	return router
}

// Middleware to check if the user is already created
func (c *userSetupController) userSetupMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		hasAnyUser, err := models.HasAnyUsers()
		if err != nil {
			c.logger.Warn("error getting number of users", err)
		}

		if hasAnyUser {
			c.flash(w, r, enums.FlashInfo, "User is already created", true)
			http.Redirect(w, r, "/users/login", http.StatusSeeOther)
			return
		}

		// if the user is not created, then allow the user to create the user
		handler.ServeHTTP(w, r)
	})
}

func (c *userSetupController) displayCreateUser(w http.ResponseWriter, r *http.Request) {
	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Create User"
	if err := views.RenderPlainTempl(userpages.CreateUser(templateData), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template", err)
	}
}

func (c *userSetupController) handleUserCreationError(w http.ResponseWriter, r *http.Request, messages ...string) {
	for _, message := range messages {
		c.flash(w, r, enums.FlashWarning, message, true)
	}

	http.Redirect(w, r, "/setup/create-user", http.StatusSeeOther)
}

func (c *userSetupController) createUserHandler(w http.ResponseWriter, r *http.Request) {
	requiredParams := []string{"username", "password", "confirm_password"}
	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.handleUserCreationError(w, r, "Please fill all the inputs")
		return
	}

	name := "Customer"
	email := "default@example.com"
	username := strings.TrimSpace(r.PostFormValue("username"))
	password := strings.TrimSpace(r.PostFormValue("password"))
	confirmPassword := strings.TrimSpace(r.PostFormValue("confirm_password"))

	if password != confirmPassword {
		c.handleUserCreationError(w, r, "Passwords do not match. Please check the password confirmation once")
		return
	}

	d, _ := models.GetUserByUsername(username)
	if d.Username != "" {
		c.handleUserCreationError(w, r, "Username Already Exists")
		return
	}

	role, err := models.GetRoleByKeyword("customer")
	if err != nil {
		c.logger.Error("Error getting role", err)
		c.handleUserCreationError(w, r, "Unable to create user")
		return
	}
	hash, err := auth.GeneratePasswordHash(password)
	if err != nil {
		c.logger.Error("Error rendering template", err)
		c.handleUserCreationError(w, r, "Unable to create user")
		return
	}
	user := models.User{
		Name:     name,
		Username: username,
		Role:     role,
		Email:    email,
		Password: hash,
		RoleID:   role.ID,
	}

	userErr := models.SaveUser(&user)
	if userErr != nil {
		c.logger.Error("Error when saving the user", err)
		c.handleUserCreationError(w, r, "Unable to create user")
		return
	}
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
