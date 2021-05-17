package comms

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Discord struct {
	Client      *discordgo.Session
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
		return fmt.Errorf("discord connect: %w", errors.New("no bot token"))
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", authToken))
	if err != nil {
		d.Logger.Errorf("discord connect: %+v", err)
		return fmt.Errorf("discord connect: %w", err)
	}

	d.Client = discord

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
		return fmt.Errorf("discord parseCredentials decode: %w", err)
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
	m, err := d.Client.ChannelMessageSend(d.Credentials.Channel, "Tester")
	if err != nil {
		d.Logger.Errorf("discord send channelMessageSend: %+v", err)
		return fmt.Errorf("discord send channelMessageSend: %w", err)
	}

	fmt.Printf("%+v", m)

	return nil
}
