package comms

import (
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Credentials struct {
	AgentID string `json:"agent_id"`
}

type CommsPackage struct {
	AgentID      string
	Message      string
	Link         string
	TicketSystem string
}

//go:generate mockery --name=CommsSystem
type CommsSystem interface {
	Connect() error

	ParseCredentials(interface{}) error

	Send(cp CommsPackage) error
}

type Comms struct {
	Config config.Config
}

func NewComms(c config.Config) *Comms {
	return &Comms{
		Config: c,
	}
}

func (c Comms) fetchCommsCredentials(agentID string) (database.CommsCredentials, error) {
	system, err := database.NewCommsStorage(*database.New(c.Config)).FetchCredentials(agentID)
	if err != nil {
		return database.CommsCredentials{
			AgentID: agentID,
			System:  "mock",
		}, bugLog.Errorf("comms fetchCommsCredentials: %w", err)
	}

	return system, nil
}

// nolint: gocyclo
func (c Comms) fetchCommsSystem(creds database.CommsCredentials) (CommsSystem, error) {
	var cs CommsSystem
	switch creds.System {
	case "slack":
		cs = NewSlack(c.Config)
	case "ms_teams":
		return nil, bugLog.Errorf("%s not yet implemented", creds.System)
	case "discord":
		cs = NewDiscord(c.Config)
	default:
		return nil, bugLog.Errorf("comms system %s is unknown", creds.System)
	}

	return cs, nil
}

func (c Comms) CommsSend(system CommsSystem, creds database.CommsCredentials, commsPackage CommsPackage) error {
	if err := system.ParseCredentials(creds); err != nil {
		return bugLog.Errorf("commsSend parseCredentials: %w", err)
	}
	if err := system.Connect(); err != nil {
		return bugLog.Errorf("commsSend connect: %w", err)
	}
	if err := system.Send(commsPackage); err != nil {
		return bugLog.Errorf("commsSend send: %w", err)
	}

	return nil
}

func (c Comms) SendComms(commsPackage CommsPackage) error {
	creds, err := c.fetchCommsCredentials(commsPackage.AgentID)
	if err != nil {
		return bugLog.Errorf("sendComms fetchCommsCredentials: %w", err)
	}
	commsSystem, err := c.fetchCommsSystem(creds)
	if err != nil {
		return bugLog.Errorf("sendComms fetchCommsSystem: %w", err)
	}

	if err := c.CommsSend(commsSystem, creds, commsPackage); err != nil {
		return bugLog.Errorf("sendComms commsSend: %w", err)
	}

	return nil
}
