package ticketing

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"

	"go.uber.org/zap"
)

type Credentials struct {
}

type TicketID string
type Hash string
type Status string

//go:generate mockery --name=TicketingSystem
type TicketingSystem interface {
	Connect() error

	ParseCredentials(interface{}) error
	FetchTicket(hash Hash) (TicketID, error)

	FetchStatus() (Status, error)

	Create() error
	Update() error
}

type Ticketing struct {
	Config config.Config
	Logger zap.SugaredLogger
}

func NewTicketing(c config.Config, logger zap.SugaredLogger) *Ticketing {
	return &Ticketing{
		Config: c,
		Logger: logger,
	}
}

type Ticket struct {
	Level string
	Bug   string
}

func (t Ticketing) fetchSystem(agentID string) (database.TicketingCredentials, error) {
	system, err := database.NewTicketingStorage(*database.New(t.Config, &t.Logger)).FetchCredentials(agentID)
	if err != nil {
		return database.TicketingCredentials{
			System: "mock",
		}, fmt.Errorf("ticketing failed to fetch system: %w", err)
	}

	return system, nil
}

func (t Ticketing) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	var ticket Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		t.Logger.Errorf("ticket parse failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Body is missing",
		}); err != nil {
			t.Logger.Errorf("ticket parse failed json: %+v", err)
		}
		return
	}

	agentId := r.Header.Get("x-agent-id")

	system, err := t.fetchSystem(agentId)
	if err != nil {
		t.Logger.Errorf("ticket fetch system failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(struct {
			Error     string
			FullError string
		}{
			Error:     "Invalid AgentID",
			FullError: fmt.Sprintf("%+v", err),
		}); err != nil {
			t.Logger.Errorf("ticket parse failed json: %v", err)
		}
		return
	}

	if err := t.createTicket(system, ticket); err != nil {
		t.Logger.Errorf("ticket create failed: %v, %+v", err, r)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: fmt.Sprintf("ticket create failed: %+v", err),
		}); err != nil {
			t.Logger.Errorf("ticket create failed json: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (t Ticketing) createTicket(system database.TicketingCredentials, ticket Ticket) error {
	switch system.System {
	case "github":
		g := NewGithub()
		if err := g.ParseCredentials(system); err != nil {
			return fmt.Errorf("failed to parse github credentials: %w", err)
		}
		if err := g.Connect(); err != nil {
			return fmt.Errorf("failed to connect github: %w", err)
		}
		if err := g.Create(); err != nil {
			return fmt.Errorf("failed to create github issue: %w", err)
		}
	}

	return nil
}
