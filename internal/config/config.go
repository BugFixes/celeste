package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	AWSEndpoint string `env:"AWS_ENDPOINT" envDefault:"https://localhost.localstack.cloud:4566"`

	DBRegion       string `env:"DB_REGION" envDefault:"us-east-1"`
	BugsTable      string `env:"DB_BUGS_TABLE" envDefault:"bugs"`
	AccountsTable  string `env:"DB_ACCOUNTS_TABLE" envDefault:"accounts"`
	AgentsTable    string `env:"DB_AGENTS_TABLE" envDefault:"agents"`
	TicketingTable string `env:"DB_TICKETING_TABLE" envDefault:"ticketing"`
	TicketsTable   string `env:"DB_TICKETS_TABLE" envDefault:"tickets"`
	CommsTable     string `env:"DB_COMMS_TABLE" envDefault:"comms"`

	QueueName string `env:"QUEUE_NAME" envDefault:"bugs"`

	LocalPort int `env:"LOCAL_PORT" envDefault:"3000"`
}

func BuildConfig() (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("config build: %w", err)
	}

	return cfg, nil
}
