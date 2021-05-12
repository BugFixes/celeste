package ticketing

import (
	"fmt"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"

	"go.uber.org/zap"
)

type Credentials struct {
	AgentID string
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
	Level         string `json:"level"`
	Bug           string `json:"bug"`
	Raw           string `json:"raw"`
	AgentID       string `json:"agent_id"`
	Line          string `json:"line"`
	File          string `json:"file"`
	ReportedTimes int    `json:"reported_times" default:"1"`
}

func (t Ticketing) fetchSystem(agentID string) (database.TicketingCredentials, error) {
	system, err := database.NewTicketingStorage(*database.New(t.Config, &t.Logger)).FetchCredentials(agentID)
	if err != nil {
		return database.TicketingCredentials{
			AgentID: agentID,
			System:  "mock",
		}, fmt.Errorf("ticketing failed to fetch system: %w", err)
	}

	return system, nil
}

func (t Ticketing) createTicket(system database.TicketingCredentials, ticket Ticket) error {
	if ticket.ReportedTimes < 1 {
		ticket.ReportedTimes = 1
	}

	switch system.System {
	case "github":
		g := NewGithub(t.Config, t.Logger)
		if err := g.ParseCredentials(system); err != nil {
			return fmt.Errorf("failed to parse github credentials: %w", err)
		}
		if err := g.Connect(); err != nil {
			return fmt.Errorf("failed to connect github: %w", err)
		}
		if err := g.Create(ticket); err != nil {
			return fmt.Errorf("failed to create github issue: %w", err)
		}
	default:
		return fmt.Errorf("failed to find system")
	}

	return nil
}
