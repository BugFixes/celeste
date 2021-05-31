package comms

import (
  "github.com/bugfixes/celeste/internal/config"
  "github.com/bugfixes/celeste/internal/database"
  bugLog "github.com/bugfixes/go-bugfixes/logs"
  "go.uber.org/zap"
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
	Logger zap.SugaredLogger
}

func NewComms(c config.Config, logger zap.SugaredLogger) *Comms {
	return &Comms{
		Config: c,
		Logger: logger,
	}
}

func (c Comms) fetchCommsCredentials(agentID string) (database.CommsCredentials, error) {
	system, err := database.NewCommsStorage(*database.New(c.Config, &c.Logger)).FetchCredentials(agentID)
	if err != nil {
		c.Logger.Errorf("comms fetchCommsCredentials: %+v", err)
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
		cs = NewSlack(c.Config, c.Logger)
	case "ms_teams":
		return nil, bugLog.Errorf("%s not yet implemented", creds.System)
	case "discord":
		cs = NewDiscord(c.Config, c.Logger)
	default:
		return nil, bugLog.Errorf("comms system %s is unknown", creds.System)
	}

	return cs, nil
}

func (c Comms) CommsSend(system CommsSystem, creds database.CommsCredentials, commsPackage CommsPackage) error {
	if err := system.ParseCredentials(creds); err != nil {
		c.Logger.Errorf("commsSend parseCredentials: %+v", err)
		return bugLog.Errorf("commsSend parseCredentials: %w", err)
	}
	if err := system.Connect(); err != nil {
		c.Logger.Errorf("commsSend connect: %+v", err)
		return bugLog.Errorf("commsSend connect: %w", err)
	}
	if err := system.Send(commsPackage); err != nil {
		c.Logger.Errorf("commsSend send: %+v", err)
		return bugLog.Errorf("commsSend send: %w", err)
	}

	return nil
}

func (c Comms) SendComms(commsPackage CommsPackage) error {
	creds, err := c.fetchCommsCredentials(commsPackage.AgentID)
	if err != nil {
		c.Logger.Errorf("sendComms fetchCommsCredentials: %+v", err)
		return bugLog.Errorf("sendComms fetchCommsCredentials: %w", err)
	}
	commsSystem, err := c.fetchCommsSystem(creds)
	if err != nil {
		c.Logger.Errorf("sendComms fetchCommsSystem: %+v", err)
		return bugLog.Errorf("sendComms fetchCommsSystem: %w", err)
	}

	if err := c.CommsSend(commsSystem, creds, commsPackage); err != nil {
		c.Logger.Errorf("sendComms commsSend: %+v", err)
		return bugLog.Errorf("sendComms commsSend: %w", err)
	}

	return nil
}
