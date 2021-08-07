package config

import (
	"encoding/json"
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

type AWS struct {
	Region        string `env:"AWS_REGION" envDefault:"eu-west-2"`
	SecretsClient *secretsmanager.SecretsManager
}

type Authorization struct {
	JWTSecret    string
	CallbackHost string `env:"CALLBACK_HOST" envDefault:"http://localhost:3000"`
}

type Secret struct {
	Key    string
	Secret string
}

type Config struct {
	Local
	RDS
	DynamoDB
	Queues
	Authorization
	AWS

	AuthCredentials []ServiceCredential
	DateFormat      string
}

func BuildConfig() (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, bugLog.Errorf("parse: %+v", err)
	}

	sess, err := BuildSession(cfg)
	if err != nil {
		return Config{}, bugLog.Errorf("buildSession: %+v", err)
	}

	cfg.AWS.SecretsClient = secretsmanager.New(sess)

	if err := buildDatabase(&cfg); err != nil {
		return cfg, bugLog.Errorf("buildDatabase: %+v", err)
	}

	buildSystems(&cfg, "PROVIDERS_LIST")

	if err := getJWTSecret(&cfg); err != nil {
		return cfg, bugLog.Errorf("getJWTSecret: %+v", err)
	}

	cfg.DateFormat = "2006-04-02 15:04:05"

	return cfg, nil
}

func BuildSession(cfg Config) (*session.Session, error) {
	if cfg.Local.Development {
		return session.NewSession(&aws.Config{
			Region:   aws.String(cfg.AWS.Region),
			Endpoint: aws.String(cfg.AWSEndpoint),
		})
	}

	return session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWS.Region),
	})
}

func GetSecret(client *secretsmanager.SecretsManager, secret string) (string, error) {
	sec, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secret),
	})
	if err != nil {
		return "", bugLog.Errorf("getSecret: %+v", err)
	}

	return *sec.SecretString, nil
}

func GetEnv(name string) string {
	return os.Getenv(name)
}

func GetSecretEnv(client *secretsmanager.SecretsManager, secret, env string) (string, error) {
	if ret := GetEnv(env); ret != "" {
		return ret, nil
	}

	return GetSecret(client, secret)
}

func getJWTSecret(cfg *Config) error {
	jwt, err := GetSecret(cfg.AWS.SecretsClient, "JWT")
	if err != nil {
		return bugLog.Errorf("jwt_secret: %+v", err)
	}
	cfg.JWTSecret = jwt
	return nil
}

func getAuthCredentials(cfg *Config, providers string) []ServiceCredential {
	serviceCreds := []ServiceCredential{}

	services := strings.Split(providers, ",")
	for _, service := range services {
		sec, err := GetSecretEnv(cfg.AWS.SecretsClient, service, "")
		if err != nil {
			continue
		}

		secret := Secret{}
		if err := json.Unmarshal([]byte(sec), &secret); err != nil {
			bugLog.Local().Warn(err)
			continue
		}

		cred := ServiceCredential{
			Service: service,
			AuthCredential: AuthCredential{
				Key:      secret.Key,
				Secret:   secret.Secret,
				Callback: fmt.Sprintf("%s/auth/%s/callback", cfg.CallbackHost, service),
			},
		}

		serviceCreds = append(serviceCreds, cred)
	}

	return serviceCreds
}

func buildSystems(cfg *Config, systemName string) {
	if systems := os.Getenv(systemName); systems != "" {
		cfg.AuthCredentials = getAuthCredentials(cfg, systems)
	}
}

func buildDatabase(cfg *Config) error {
	r := RDS{}

	val, err := GetSecretEnv(cfg.AWS.SecretsClient, "RDSPassword", "DB_PASSWORD")
	if err != nil {
		return bugLog.Errorf("password: %+v", err)
	}
	r.Password = val

	val, err = GetSecretEnv(cfg.AWS.SecretsClient, "RDSUsername", "DB_USERNAME")
	if err != nil {
		return bugLog.Errorf("password: %+v", err)
	}
	r.Username = val

	val, err = GetSecretEnv(cfg.AWS.SecretsClient, "RDSHostname", "DB_HOSTNAME")
	if err != nil {
		return bugLog.Errorf("password: %+v", err)
	}
	r.Hostname = val

	val, err = GetSecretEnv(cfg.AWS.SecretsClient, "RDSPort", "DB_PORT")
	if err != nil {
		return bugLog.Errorf("password: %+v", err)
	}
	r.Port = val

	val, err = GetSecretEnv(cfg.AWS.SecretsClient, "RDSDatabase", "DB_DATABASE")
	if err != nil {
		return bugLog.Errorf("password: %+v", err)
	}
	r.Database = val

	cfg.RDS = r

	return err
}
