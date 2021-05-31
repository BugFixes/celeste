package comms

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Discord struct {
	BotAuth string

	Context     context.Context
	Credentials DiscordCredentials
	Config      config.Config
	Logger      zap.SugaredLogger
}

type DiscordCredentials struct {
	Channel string `json:"channel"`
	Credentials
}

func NewDiscord(c config.Config, logger zap.SugaredLogger) *Discord {
	return &Discord{
		Context: context.Background(),
		Config:  c,
		Logger:  logger,
	}
}

func (d *Discord) Connect() error {
	authToken := os.Getenv("DISCORD_BOT_TOKEN")
	if authToken == "" {
		d.Logger.Errorf("discord connect: %+v", errors.New("no bot token"))
		return bugLog.Errorf("discord connect: %w", errors.New("no bot token"))
	}

	d.BotAuth = authToken

	return nil
}

func (d *Discord) ParseCredentials(creds interface{}) error {
	type cc struct {
		AgentID      string `json:"agent_id"`
		System       string `json:"system"`
		CommsDetails struct {
			Channel string `json:"channel"`
		} `json:"comms_details"`
	}

	discordCreds := cc{}
	if err := mapstructure.Decode(creds, &discordCreds); err != nil {
		d.Logger.Errorf("discord parseCredentials decode: %+v", err)
		return bugLog.Errorf("discord parseCredentials decode: %w", err)
	}

	d.Credentials = DiscordCredentials{
		Channel: discordCreds.CommsDetails.Channel,
		Credentials: Credentials{
			AgentID: discordCreds.AgentID,
		},
	}

	return nil
}

func (d *Discord) Send(commsPackage CommsPackage) error {
	title := fmt.Sprintf("A new ticket has been added to %s by BugFix.es", commsPackage.TicketSystem)
	embed := discord.Embed{
		Title: "Ticket Link",
		URL:   commsPackage.Link,
	}

	g, err := gateway.NewGateway(d.BotAuth)
	if err != nil {
		d.Logger.Errorf("discord send newGateway: %*v", err)
		return bugLog.Errorf("discord send newGateway: %w", err)
	}
	g.AddIntents(gateway.IntentGuildMessages)
	if err := g.OpenContext(d.Context); err != nil {
		d.Logger.Errorf("discord send openContext: %+v", err)
		return bugLog.Errorf("discord send openContext: %w", err)
	}

	c := api.NewClient(fmt.Sprintf("Bot %s", d.BotAuth)).WithContext(d.Context)
	snow, err := discord.ParseSnowflake(d.Credentials.Channel)
	if err != nil {
		d.Logger.Errorf("discord send parseSnowFlake: %+v", err)
		return bugLog.Errorf("discord send parseSnowFlake: %w", err)
	}
	m, err := c.SendMessage(discord.ChannelID(snow), title, &embed)
	if err != nil {
		d.Logger.Errorf("discord send sendMessage: %+v", err)
		return bugLog.Errorf("discord send sendMessage: %w", err)
	}

	bugLog.Infof("%+v", m)

	return nil
}
