package config

import (
	"fmt"
	"os"
	"strings"

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

type ServiceCredential struct {
	Service string
	AuthCredential
}

type AuthCredential struct {
	Key      string
	Secret   string
	Callback string
}

type DynamoDB struct {
	BugsTable      string `env:"DB_BUGS_TABLE" envDefault:"bugs"`
	AccountsTable  string `env:"DB_ACCOUNTS_TABLE" envDefault:"accounts"`
	AgentsTable    string `env:"DB_AGENTS_TABLE" envDefault:"agents"`
	TicketingTable string `env:"DB_TICKETING_TABLE" envDefault:"ticketing"`
	TicketsTable   string `env:"DB_TICKETS_TABLE" envDefault:"tickets"`
	CommsTable     string `env:"DB_COMMS_TABLE" envDefault:"comms"`
	LogsTable      string `env:"DB_LOGS_TABLE" envDefault:"logs"`
}

type Local struct {
	KeepLocal   bool   `env:"LOCAL_ONLY" envDefault:"false"`
	Development bool   `env:"DEVELOPMENT" envDefault:"true"`
	AWSEndpoint string `env:"AWS_ENDPOINT" envDefault:"https://localhost.localstack.cloud:4566"`
	Port        int    `env:"LOCAL_PORT" envDefault:"3000"`
}

type Queues struct {
	Name       string `env:"QUEUE_NAME" envDefault:"bugs"`
	DeadLetter string `env:"QUEUE_DEADLETTER_NAME" envDefault:"deadletter"`
}

type Authorization struct {
	JWTSecret    string
	CallbackHost string `env:"CALLBACK_HOST" envDefault:"http://localhost:3000"`
}

type Config struct {
	Local
	RDS
	DynamoDB
	Queues
	Authorization

	AWSRegion string `env:"AWS_REGION" envDefault:"eu-west-2"`

	AuthCredentials []ServiceCredential
	DateFormat      string
}

func BuildConfig() (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, bugLog.Errorf("parse: %w", err)
	}

	sess, err := BuildSession(cfg)
	if err != nil {
		return Config{}, bugLog.Errorf("buildSession: %w", err)
	}

	secretClient := secretsmanager.New(sess)

	if err := buildDatabase(&cfg, secretClient); err != nil {
		return cfg, bugLog.Errorf("buildDatabase: %w", err)
	}

	buildProviders(&cfg, secretClient)

	if err := getJWTSecret(&cfg, secretClient); err != nil {
		return cfg, bugLog.Errorf("getJWTSecret: %w", err)
	}

	cfg.DateFormat = "2006-04-02 15:04:05"

	return cfg, nil
}

func BuildSession(cfg Config) (*session.Session, error) {
	if cfg.Local.Development {
		return session.NewSession(&aws.Config{
			Region:   aws.String(cfg.AWSRegion),
			Endpoint: aws.String(cfg.AWSEndpoint),
		})
	}

	return session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
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

func getJWTSecret(cfg *Config, client *secretsmanager.SecretsManager) error {
	jwt, err := getSecret(client, "jwt_secret")
	if err != nil {
		return bugLog.Errorf("jwt_secret: %w", err)
	}
	cfg.JWTSecret = jwt
	return nil
}

func getAuthCredentials(cfg *Config, providers string, client *secretsmanager.SecretsManager) []ServiceCredential {
	serviceCreds := []ServiceCredential{}

	services := strings.Split(providers, ",")
	for _, service := range services {
		key, err := getAuthSecret(client, service, "key")
		if err != nil {
			continue
		}
		sec, err := getAuthSecret(client, service, "secret")
		if err != nil {
			continue
		}
		cred := ServiceCredential{
			Service: service,
			AuthCredential: AuthCredential{
				Key:      key,
				Secret:   sec,
				Callback: fmt.Sprintf("%s/auth/%s/callback", cfg.CallbackHost, service),
			},
		}

		serviceCreds = append(serviceCreds, cred)
	}

	return serviceCreds
}

func getAuthSecret(client *secretsmanager.SecretsManager, service, secret string) (string, error) {
	sec, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(fmt.Sprintf("%s_%s", service, secret)),
	})
	if err != nil {
		return "", bugLog.Errorf("getAuthSecret: %s_%s, %w", service, secret, err)
	}

	return *sec.SecretString, nil
}

func buildProviders(cfg *Config, client *secretsmanager.SecretsManager) {
	if providers := os.Getenv("PROVIDERS_LIST"); providers != "" {
		cfg.AuthCredentials = getAuthCredentials(cfg, providers, client)
	}
}

func buildDatabase(cfg *Config, client *secretsmanager.SecretsManager) error {
	r := RDS{}

	val, err := getSecret(client, "RDSPassword")
	if err != nil {
		return bugLog.Errorf("password: %w", err)
	}
	r.Password = val

	val, err = getSecret(client, "RDSUsername")
	if err != nil {
		return bugLog.Errorf("password: %w", err)
	}
	r.Username = val

	val, err = getSecret(client, "RDSHostname")
	if err != nil {
		return bugLog.Errorf("password: %w", err)
	}
	r.Hostname = val

	val, err = getSecret(client, "RDSPort")
	if err != nil {
		return bugLog.Errorf("password: %w", err)
	}
	r.Port = val

	val, err = getSecret(client, "RDSDatabase")
	if err != nil {
		return bugLog.Errorf("password: %w", err)
	}
	r.Database = val

	cfg.RDS = r

	return err
}
