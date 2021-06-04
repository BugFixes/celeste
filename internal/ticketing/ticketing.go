package ticketing

import (
	"crypto/sha256"
	"fmt"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Credentials struct {
	AgentID string
}

type TicketID string
type Hash string
type Status string

type TicketTemplate struct {
	Title  string      `json:"title"`
	Body   interface{} `json:"body"`
	Labels []string    `json:"labels"`
	Level  string      `json:"level"`
}

const (
	firstReport = "first report"
	multiReport = "multiple reports"
)

//go:generate mockery --name=TicketingSystem
type TicketingSystem interface {
	Connect() error

	ParseCredentials(interface{}) error
	FetchRemoteTicket(interface{}) (Ticket, error)

	Create(*Ticket) error
	Update(*Ticket) error
	Fetch(*Ticket) error

	GenerateTemplate(*Ticket) (TicketTemplate, error)
	TicketExists(*Ticket) (bool, database.TicketDetails, error)
}

type Ticketing struct {
	Config config.Config
}

func NewTicketing(c config.Config) *Ticketing {
	return &Ticketing{
		Config: c,
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

	RemoteID      string      `json:"remote_id"`
	RemoteDetails interface{} `json:"remote_details"`
	Hash          Hash        `json:"hash"`
	State         string      `json:"state"`
	RemoteLink    string      `json:"remote_link"`
	RemoteSystem  string      `json:"remote_system"`
}

func (t Ticketing) fetchTicketingCredentials(agentID string) (database.TicketingCredentials, error) {
	system, err := database.NewTicketingStorage(*database.New(t.Config)).FetchCredentials(agentID)
	if err != nil {
		return database.TicketingCredentials{
			AgentID: agentID,
			System:  "mock",
		}, bugLog.Errorf("ticketing failed to fetch system: %w", err)
	}

	return system, nil
}

// nolint: gocyclo
func (t Ticketing) fetchTicketSystem(creds database.TicketingCredentials) (TicketingSystem, error) {
	var ts TicketingSystem

	switch creds.System {
	case "github":
		ts = NewGithub(t.Config)
	case "jira":
		ts = NewJira(t.Config)
	case "trac":
	case "youtrack":
	case "proofhub":
	case "backlog":
	case "orapm":
	case "bugzilla":
	case "asana":
		return nil, bugLog.Errorf("%s not yet implemented", creds.System)
	default:
		return nil, bugLog.Errorf("ticket system %s is unknown", creds.System)
	}

	return ts, nil
}

func (t Ticketing) TicketCreate(system TicketingSystem, creds database.TicketingCredentials, ticket *Ticket) error {
	ticket.RemoteSystem = creds.System

	if err := system.ParseCredentials(creds); err != nil {
		return bugLog.Errorf("ticketCreate parseCredentials: %w", err)
	}
	if err := system.Connect(); err != nil {
		return bugLog.Errorf("ticketCreate connect: %w", err)
	}
	if err := system.Create(ticket); err != nil {
		return bugLog.Errorf("ticketCreate create: %w", err)
	}
	return nil
}

func (t Ticketing) CreateTicket(ticket *Ticket) error {
	ticketSystemCredentials, err := t.fetchTicketingCredentials(ticket.AgentID)
	if err != nil {
		return bugLog.Errorf("createTicket fetchSystem failed: %w", err)
	}

	ticketSystem, err := t.fetchTicketSystem(ticketSystemCredentials)
	if err != nil {
		return bugLog.Errorf("createTicket fetchTicketSystem: %w", err)
	}

	if err := t.TicketCreate(ticketSystem, ticketSystemCredentials, ticket); err != nil {
		return bugLog.Errorf("createTicket ticketCreate: %w", err)
	}

	return nil
}

func GenerateHash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}
