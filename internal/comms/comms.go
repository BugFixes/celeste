package comms

import (
	"github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Credentials struct {
	AgentID string `json:"agent_id"`
}

type CommsPackage struct {
	Agent        agent.Agent
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

func (c Comms) fetchCommsCredentials(a agent.Agent) (CommsCredentials, error) {
	system, err := NewCommsStorage(c.Config).FetchCredentials(a)
	if err != nil {
		return CommsCredentials{
			Agent:  a,
			System: "mock",
		}, bugLog.Errorf("comms fetchCommsCredentials: %+v", err)
	}

	return system, nil
}

// nolint: gocyclo
func (c Comms) fetchCommsSystem(creds CommsCredentials) (CommsSystem, error) {
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

func (c Comms) CommsSend(system CommsSystem, creds CommsCredentials, commsPackage CommsPackage) error {
	if err := system.ParseCredentials(creds); err != nil {
		return bugLog.Errorf("commsSend parseCredentials: %+v", err)
	}
	if err := system.Connect(); err != nil {
		return bugLog.Errorf("commsSend connect: %+v", err)
	}
	if err := system.Send(commsPackage); err != nil {
		return bugLog.Errorf("commsSend send: %+v", err)
	}

	return nil
}

func (c Comms) SendComms(commsPackage CommsPackage) error {
	creds, err := c.fetchCommsCredentials(commsPackage.Agent)
	if err != nil {
		return bugLog.Errorf("sendComms fetchCommsCredentials: %+v", err)
	}
	commsSystem, err := c.fetchCommsSystem(creds)
	if err != nil {
		return bugLog.Errorf("sendComms fetchCommsSystem: %+v", err)
	}

	if err := c.CommsSend(commsSystem, creds, commsPackage); err != nil {
		return bugLog.Errorf("sendComms commsSend: %+v", err)
	}

	return nil
}
