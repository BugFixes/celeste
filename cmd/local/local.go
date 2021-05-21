package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bugfixes/celeste/internal/celeste/account"
	"github.com/bugfixes/celeste/internal/celeste/bug"
	"github.com/bugfixes/celeste/internal/comms"
	"github.com/bugfixes/celeste/internal/ticketing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/celeste"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/go-bugfixes"
)

func main() {
	// Logger
	logger, err := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	if err != nil {
		fmt.Printf("zap failed to start: %v", err)
		return
	}
	sugar := logger.Sugar()
	sugar.Infow("Starting Celeste")

	// Config
	cfg, err := config.BuildConfig()
	if err != nil {
		fmt.Printf("failed to build config: %v", err)
		return
	}

	// Celeste
	c := celeste.Celeste{
		Config: cfg,
		Logger: sugar,
	}

	if err := route(c); err != nil {
		fmt.Printf("failed to route: %v", err)
		return
	}
}

func route(c celeste.Celeste) error {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)
	r.Use(bugfixes.Logger)
	r.Use(bugfixes.Recoverer)

	// Account
	r.Route("/account", func(r chi.Router) {
		r.Post("/", account.NewHTTPRequest(c.Config, *c.Logger).CreateHandler)
		r.Delete("/", account.NewHTTPRequest(c.Config, *c.Logger).DeleteHandler)

		r.Post("/login", account.NewHTTPRequest(c.Config, *c.Logger).LoginHandler)
	})

	// Agent
	r.Route("/agent", func(r chi.Router) {

	})

	// Bug
	r.Route("/bug", func(r chi.Router) {
		r.Post("/", bug.NewBug(c.Config, *c.Logger).BugHandler)
	})

	// Comms
	r.Route("/comms", func(r chi.Router) {
		r.Post("/", comms.NewCommunication(c.Config, *c.Logger).CreateCommsHandler)
		r.Put("/", comms.NewCommunication(c.Config, *c.Logger).AttachCommsHandler)
		r.Patch("/", comms.NewCommunication(c.Config, *c.Logger).DetachCommsHandler)
		r.Delete("/", comms.NewCommunication(c.Config, *c.Logger).DeleteCommsHandler)
		r.Get("/", comms.NewCommunication(c.Config, *c.Logger).ListCommsHandler)
	})

	// Ticket
	r.Route("/ticket", func(r chi.Router) {
		r.Post("/", ticketing.NewTicketing(c.Config, *c.Logger).CreateTicketHandler)
	})

	fmt.Printf("listening on port: %d\n", c.Config.LocalPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Config.LocalPort), r)
	if err != nil {
		return fmt.Errorf("failed to start port: %w", err)
	}

	return nil
}
