package ticketing

import (
	"crypto/sha256"
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
	FetchRemoteTicket(Hash) (Ticket, error)

	Create(Ticket) error
	Update(Ticket) error
	Fetch(Ticket) (Ticket, error)
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
	LevelNumber   string `json:"level_number"`
	Bug           string `json:"bug"`
	Raw           string `json:"raw"`
	AgentID       string `json:"agent_id"`
	Line          string `json:"line"`
	File          string `json:"file"`
	TimesReported int    `json:"times_reported" default:"1"`
}

func (t Ticketing) fetchTicketingCredentials(agentID string) (database.TicketingCredentials, error) {
	system, err := database.NewTicketingStorage(*database.New(t.Config, &t.Logger)).FetchCredentials(agentID)
	if err != nil {
		return database.TicketingCredentials{
			AgentID: agentID,
			System:  "mock",
		}, fmt.Errorf("ticketing failed to fetch system: %w", err)
	}

	return system, nil
}

func (t Ticketing) fetchTicketSystem(creds database.TicketingCredentials) (TicketingSystem, error) {
	var ts TicketingSystem

	switch creds.System {
	case "github":
		ts = NewGithub(t.Config, t.Logger)
	case "jira":
		// TODO jira
		return nil, fmt.Errorf("not yet implemented")
	default:
		return nil, fmt.Errorf("failed to find system")
	}

	return ts, nil
}

func (t Ticketing) TicketCreate(system TicketingSystem, creds database.TicketingCredentials, ticket Ticket) error {
	if err := system.ParseCredentials(creds); err != nil {
		return fmt.Errorf("ticket create parse credentials: %w", err)
	}
	if err := system.Connect(); err != nil {
		return fmt.Errorf("ticket create connect: %w", err)
	}
	if err := system.Create(ticket); err != nil {
		return fmt.Errorf("ticket create create: %w", err)
	}
	return nil
}

func (t Ticketing) CreateTicket(ticket Ticket) error {
	system, err := t.fetchTicketingCredentials(ticket.AgentID)
	if err != nil {
		return fmt.Errorf("createTicket fetchSystem failed: %w", err)
	}

	ts, err := t.fetchTicketSystem(system)
	if err != nil {
		return fmt.Errorf("createTicket fetchTicketSystem: %w", err)
	}

	if err := t.TicketCreate(ts, system, ticket); err != nil {
		return fmt.Errorf("createTicket ticketCreate: %w", err)
	}

	return nil
}

func GenerateHash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}
