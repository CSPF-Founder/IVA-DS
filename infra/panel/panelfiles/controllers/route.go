package controllers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/CSPF-Founder/iva/panel/config"
	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/db"
	"github.com/CSPF-Founder/iva/panel/frontend"
	"github.com/CSPF-Founder/iva/panel/internal/httpclient"
	"github.com/CSPF-Founder/iva/panel/internal/sessions"
	"github.com/CSPF-Founder/iva/panel/logger"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	server     *http.Server
	config     *config.Config
	logger     *logger.Logger
	session    *sessions.SessionManager
	httpClient httpclient.HttpClient
}

// Change Configuration accordingly
var defaultTLSConfig = &tls.Config{
	PreferServerCipherSuites: true,
	CurvePreferences: []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	},
	MinVersion: tls.VersionTLS12,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

		// Kept for backwards compatibility with some clients
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	},
}

// ServerOption is a functional option that is used to configure the
type ServerOption func(*App)

// NewApp returns a new instance of the app with
// provided options applied.
func NewApp(cfg *config.Config, appLogger *logger.Logger, options ...ServerOption) *App {
	defaultServer := &http.Server{
		ReadTimeout:  config.ServerReadTimeout,
		WriteTimeout: config.ServerWriteTimeout,
		Addr:         cfg.ServerConf.ServerAddress,
	}

	sessionManager := sessions.SetupSession(cfg)
	app := &App{
		server:  defaultServer,
		config:  cfg,
		logger:  appLogger,
		session: sessionManager,
	}
	for _, opt := range options {
		opt(app)
	}

	app.httpClient = &http.Client{
		Timeout: config.HttpClientTimeout,
	}
	return app
}

func (app *App) SetupDB() {
	err := db.Setup(app.config, app.session)
	if err != nil {
		app.logger.Fatal("Error setting up models", err)
	}

}

// Start launches the server, listening on the configured address.
func (app *App) StartServer() {
	// Use Tls if configured
	if app.config.ServerConf.UseTLS {
		app.server.TLSConfig = defaultTLSConfig

		app.logger.Info("Creating new self-signed certificate")
		err := utils.CheckAndCreateSSL(app.config.ServerConf.CertPath, app.config.ServerConf.KeyPath)
		if err != nil {
			app.logger.Fatal("Error creating SSL Certificates: ", err)
			return
		}

		app.logger.Info("TLS Certificate Generation complete")

		app.logger.Info(fmt.Sprintf("Starting server at https://%s", app.config.ServerConf.ServerAddress))
		err = app.server.ListenAndServeTLS(app.config.ServerConf.CertPath, app.config.ServerConf.KeyPath)
		if err != nil {
			app.logger.Fatal("Error starting server: ", err)
		}
	}
	// If TLS isn't configured, just listen on HTTP
	app.logger.Info(fmt.Sprintf("Starting server at http://%s", app.config.ServerConf.ServerAddress))
	err := app.server.ListenAndServe()
	if err != nil {
		app.logger.Fatal("Error starting server: ", err)
	}
}

// Shutdown attempts to gracefully shutdown the server.
func (app *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return app.server.Shutdown(ctx)
}

func (app *App) SetupRoutes() {

	dataRepo, err := db.SetupMongo(context.TODO(), app.config)
	if err != nil {
		app.logger.Fatal("Error setting up Mongo", err)
	}

	router := chi.NewRouter() // Initialize Chi router

	// r.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	errHandlers := newErrorController(app)

	// CSRF Validation Middleware
	router.Use(mid.CSRFMiddleware(
		app.session,
		http.HandlerFunc(errHandlers.csrfErrorHandler),
		[]string{"/api"}),
	)

	usec := newUserSetupController(app)
	uc := newUserController(app)
	update := newUpdateController(app)

	target := newTargetController(app, dataRepo.DB, dataRepo.Target, dataRepo.ScanResult)
	bugTrack := newBugTrackController(
		app,
		dataRepo.DB,
		dataRepo.ScanResult,
		dataRepo.Target,
	)

	// Middlewares
	router.NotFound(errHandlers.notFoundHandler)
	// Setup logging
	router.Use(mid.LoggingMiddleware(app.logger))

	router.Get("/", mid.Use(app.handleHomePage))
	router.Mount("/users", uc.registerRoutes())
	router.Mount("/setup", usec.registerRoutes())
	router.Mount("/targets", target.registerRoutes())
	router.Mount("/update", update.registerRoutes())
	router.Mount("/bug-track", bugTrack.registerRoutes())

	// Embedded static file serving
	// fileServer := http.FileServer(http.FS(frontend.FileSystem))
	fileServer := noDirListingFileServer(http.FS(frontend.FileSystem))
	router.Handle("/static/*", fileServer)

	// External static files that will be later mounted via docker
	// externalFS := http.FileServer(http.Dir("./frontend/external"))
	externalFS := noDirListingFileServer(http.Dir("./frontend/external"))
	router.Handle("/external/*", http.StripPrefix("/external/", externalFS))

	routeHandler := mid.Use(
		router.ServeHTTP,
		mid.GetContext(app.session),
		mid.ApplySecurityHeaders,
	)
	app.server.Handler = app.session.LoadAndSave(routeHandler)
}

func (c *App) handleHomePage(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user")
	if user == nil {
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
	}
	http.Redirect(w, r, "/targets/add", http.StatusSeeOther)
}

// Custom file server that disables directory listing
func noDirListingFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Open the requested file or directory
		f, err := fs.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		// Check if the path is a directory
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			http.NotFound(w, r) // Return 404 for directories
			return
		}

		// Serve the file if it's not a directory
		http.FileServer(fs).ServeHTTP(w, r)
	})
}
