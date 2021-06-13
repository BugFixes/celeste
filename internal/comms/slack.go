package comms

import (
	"context"
	"errors"
	"fmt"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/mitchellh/mapstructure"
	"github.com/slack-go/slack"
)

type Slack struct {
	Client *slack.Client

	Context     context.Context
	Credentials SlackCredentials
	Config      config.Config
}

type SlackCredentials struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
	Credentials
}

func NewSlack(c config.Config) *Slack {
	return &Slack{
		Context: context.Background(),
		Config:  c,
	}
}

func (s *Slack) Connect() error {
	authToken := s.Credentials.Token
	if authToken == "" {
		return bugLog.Errorf("slack connect: %w", errors.New("no bot token"))
	}
	s.Client = slack.New(authToken)

	return nil
}

func (s *Slack) ParseCredentials(creds interface{}) error {
	type sc struct {
		AgentID      string `json:"agent_id"`
		System       string `json:"system"`
		CommsDetails struct {
			Channel string `json:"channel"`
			Token   string `json:"token"`
		} `json:"comms_details"`
	}

	slackCreds := sc{}
	if err := mapstructure.Decode(creds, &slackCreds); err != nil {
		return bugLog.Errorf("slack parseCredentials decode: %w", err)
	}

	s.Credentials = SlackCredentials{
		Channel: slackCreds.CommsDetails.Channel,
		Token:   slackCreds.CommsDetails.Token,
		Credentials: Credentials{
			AgentID: slackCreds.AgentID,
		},
	}

	return nil
}

func (s *Slack) Send(commsPackage CommsPackage) error {
	title := slack.MsgOptionText(
		fmt.Sprintf("A new ticket has been added to %s by BugFix.es", commsPackage.TicketSystem),
		false)
	message := slack.MsgOptionAttachments(slack.Attachment{
		Text: fmt.Sprintf("Ticket Link\n %s", commsPackage.Link),
	})

	if _, _, err := s.Client.PostMessageContext(
		s.Context,
		s.Credentials.Channel,
		title,
		message); err != nil {
		return bugLog.Errorf("slack send postMessage: %w", err)
	}

	return nil
}
