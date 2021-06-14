package main

import (
	"fmt"
	"net/http"
	"time"

  "github.com/bugfixes/celeste/internal/auth"
  "github.com/bugfixes/celeste/internal/celeste"
	"github.com/bugfixes/celeste/internal/celeste/account"
	"github.com/bugfixes/celeste/internal/celeste/bug"
	"github.com/bugfixes/celeste/internal/comms"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/ticketing"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugfixes "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	bugLog.Local().Info("Starting Celeste")

	// Config
	cfg, err := config.BuildConfig()
	if err != nil {
		_ = bugLog.Errorf("buildConfig: %v", err)
		return
	}

	// Celeste
	c := celeste.Celeste{
		Config: cfg,
	}

	if err := route(c); err != nil {
		_ = bugLog.Errorf("route: %v", err)
		return
	}
}

func route(c celeste.Celeste) error {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)
	r.Use(bugfixes.BugFixes)

	// Account
	r.Route("/account", func(r chi.Router) {
		r.Post("/", account.NewHTTPRequest(c.Config).CreateHandler)
		r.Delete("/", account.NewHTTPRequest(c.Config).DeleteHandler)

		r.Post("/login", account.NewHTTPRequest(c.Config).LoginHandler)
	})

	// Auth
	r.Route("/auth", func(r chi.Router) {
	  r.Get("/{provider}/callback", auth.NewAuth(c.Config).CallbackHandler)
	  r.Get("/logout/{provider}", auth.NewAuth(c.Config).LogoutHandler)
	  r.Get("/{provider}", auth.NewAuth(c.Config).AuthHandler)
  })

	// Agent
	r.Route("/agent", func(r chi.Router) {

	})

	// Logs
	r.Route("/log", func(r chi.Router) {
		r.Post("/", bug.NewLog(c.Config).LogHandler)
	})

	// Bug
	r.Route("/bug", func(r chi.Router) {
		r.Post("/", bug.NewBug(c.Config).BugHandler)
	})

	// Comms
	r.Route("/comms", func(r chi.Router) {
		r.Post("/", comms.NewCommunication(c.Config).CreateCommsHandler)
		r.Put("/", comms.NewCommunication(c.Config).AttachCommsHandler)
		r.Patch("/", comms.NewCommunication(c.Config).DetachCommsHandler)
		r.Delete("/", comms.NewCommunication(c.Config).DeleteCommsHandler)
		r.Get("/", comms.NewCommunication(c.Config).ListCommsHandler)
	})

	// Ticket
	r.Route("/ticket", func(r chi.Router) {
		r.Post("/", ticketing.NewTicketing(c.Config).CreateTicketHandler)
	})

	bugLog.Local().Infof("listening on port: %d\n", c.Config.LocalPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Config.LocalPort), r)
	if err != nil {
		return bugLog.Errorf("port: %w", err)
	}

	return nil
}
