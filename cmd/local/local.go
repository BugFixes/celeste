package main

import (
	"fmt"
	"net/http"
	"time"

	account2 "github.com/bugfixes/celeste/internal/account"
	"github.com/bugfixes/celeste/internal/auth"
	bug2 "github.com/bugfixes/celeste/internal/bug"
	"github.com/bugfixes/celeste/internal/comms"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/frontend"
	"github.com/bugfixes/celeste/internal/handler"
	"github.com/bugfixes/celeste/internal/ticketing"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugfixes "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/mux"
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
	c := handler.Celeste{
		Config: cfg,
	}

	if err := route(c); err != nil {
		_ = bugLog.Errorf("route: %v", err)
		return
	}
}

func route(c handler.Celeste) error {
	r := mux.NewRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)
	r.Use(bugfixes.BugFixes)

	// Auth
	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/{provider}", auth.NewAuth(c.Config).AuthHandler)
	s.HandleFunc("/{provider}/callback", auth.NewAuth(c.Config).CallbackHandler)
	s.HandleFunc("/logout/{provider}", auth.NewAuth(c.Config).LogoutHandler)

	// Account
	s = r.PathPrefix("account").Subrouter()
	s.HandleFunc("/", account2.NewHTTPRequest(c.Config).CreateHandler).Methods("POST")
	s.HandleFunc("/", account2.NewHTTPRequest(c.Config).DeleteHandler).Methods("DELETE")
	s.HandleFunc("/login", account2.NewHTTPRequest(c.Config).LoginHandler).Methods("POST")

	// Agent
	// TODO: Add agent
	// s = r.PathPrefix("/agent").Subrouter()

	// Logs
	s = r.PathPrefix("/log").Subrouter()
	s.HandleFunc("/", bug2.NewLog(c.Config).LogHandler).Methods("POST")

	// Bug
	s = r.PathPrefix("/bug").Subrouter()
	s.HandleFunc("/", bug2.NewBug(c.Config).BugHandler).Methods("POST")

	// Comms
	s = r.PathPrefix("/comms").Subrouter()
	s.HandleFunc("/", comms.NewCommunication(c.Config).CreateCommsHandler).Methods("POST")
	s.HandleFunc("/", comms.NewCommunication(c.Config).AttachCommsHandler).Methods("PUT")
	s.HandleFunc("/", comms.NewCommunication(c.Config).DetachCommsHandler).Methods("PATCH")
	s.HandleFunc("/", comms.NewCommunication(c.Config).DeleteCommsHandler).Methods("DELETE")
	s.HandleFunc("/", comms.NewCommunication(c.Config).ListCommsHandler).Methods("GET")

	// Ticket
	s = r.PathPrefix("/ticket").Subrouter()
	s.HandleFunc("/", ticketing.NewTicketing(c.Config).CreateTicketHandler).Methods("POST")

	// Frontend
	s = r.PathPrefix("/fe").Subrouter()
	s.HandleFunc("/r", frontend.NewFrontend(c.Config).RegisterHandler).Methods("POST")
	s.HandleFunc("/d", frontend.NewFrontend(c.Config).DetailsHandler).Methods("GET")

	bugLog.Local().Infof("listening on port: %d\n", c.Config.LocalPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Config.LocalPort), r)
	if err != nil {
		return bugLog.Errorf("port: %w", err)
	}

	return nil
}
