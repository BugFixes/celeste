package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v6"
)

type RDS struct {
	Username string
	Password string
	Hostname string
	Port     string
	Database string
}

type Config struct {
	AWSEndpoint string `env:"AWS_ENDPOINT" envDefault:"https://localhost.localstack.cloud:4566"`

	DBRegion       string `env:"DB_REGION" envDefault:"us-east-1"`
	BugsTable      string `env:"DB_BUGS_TABLE" envDefault:"bugs"`
	AccountsTable  string `env:"DB_ACCOUNTS_TABLE" envDefault:"accounts"`
	AgentsTable    string `env:"DB_AGENTS_TABLE" envDefault:"agents"`
	TicketingTable string `env:"DB_TICKETING_TABLE" envDefault:"ticketing"`
	TicketsTable   string `env:"DB_TICKETS_TABLE" envDefault:"tickets"`
	CommsTable     string `env:"DB_COMMS_TABLE" envDefault:"comms"`
	LogsTable      string `env:"DB_LOGS_TABLE" envDefault:"logs"`
	RDS

	QueueName string `env:"QUEUE_NAME" envDefault:"bugs"`

	LocalPort int `env:"LOCAL_PORT" envDefault:"3000"`
}

func BuildConfig() (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, bugLog.Errorf("parse: %w", err)
	}

	rds, err := buildDatabase(cfg)
	if err != nil {
		return cfg, bugLog.Errorf("buildDatabase: %w", err)
	}
	cfg.RDS = rds

	return cfg, nil
}

func buildDatabase(cfg Config) (RDS, error) {
	r := RDS{}

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.DBRegion),
		Endpoint: aws.String(cfg.AWSEndpoint),
	})
	if err != nil {
		return r, bugLog.Errorf("session: %w", err)
	}
	client := secretsmanager.New(sess)

	if r.Password, err = getSecret(client, "RDSPassword"); err != nil {
		return r, bugLog.Errorf("password: %w", err)
	}

	if r.Username, err = getSecret(client, "RDSUsername"); err != nil {
		return r, bugLog.Errorf("username: %w", err)
	}

	if r.Hostname, err = getSecret(client, "RDSHostname"); err != nil {
		return r, bugLog.Errorf("hostname: %w", err)
	}

	if r.Port, err = getSecret(client, "RDSPort"); err != nil {
		return r, bugLog.Errorf("port: %w", err)
	}

	if r.Database, err = getSecret(client, "RDSDatabase"); err != nil {
		return r, bugLog.Errorf("database: %w", err)
	}

	return r, nil
}

func getSecret(client *secretsmanager.SecretsManager, secret string) (string, error) {
	sec, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secret),
	})
	if err != nil {
		return "", bugLog.Errorf("getSecret: %w", err)
	}

	return *sec.SecretString, nil
}
