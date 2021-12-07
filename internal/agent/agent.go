package agent

import (
	"context"
	"fmt"

	"github.com/bugfixes/celeste/internal/account"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

type Credentials struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type Agent struct {
	ID   int
	UUID string `json:"id"`
	Name string `json:"name"`

	Credentials
	account.Account
}

  type AgentClient struct {
    Config  config.Config
    Context context.Context
  }

//go:generate mockery --name=Agents
type Agents interface {
	Create() (*Agent, error)
	Delete(a Agent) error
}

func NewAgent(c config.Config) *AgentClient {
	return &AgentClient{
		Config:  c,
		Context: context.Background(),
	}
}

func (ac AgentClient) getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		ac.Context,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			ac.Config.RDS.Username,
			ac.Config.RDS.Password,
			ac.Config.RDS.Hostname,
			ac.Config.RDS.Port,
			ac.Config.RDS.Database))
	if err != nil {
		return nil, bugLog.Errorf("getConnnection: %+v", err)
	}

	return conn, nil
}

func (ac AgentClient) Find(a *Agent) error {
	conn, err := ac.getConnection()
	if err != nil {
		return bugLog.Errorf("find: %+v", err)
	}

	if err := conn.QueryRow(ac.Context,
		"SELECT id FROM agent WHERE key = $1 AND secret = $2 LIMIT 1",
		a.Key,
		a.Secret).Scan(&a.ID); err != nil {
		return bugLog.Errorf("find: %+v", err)
	}

	return nil
}

// func NewAgent(name string, account account2.Account) *Agent {
// 	return &Agent{
// 		Name:    name,
// 		Account: account,
// 	}
// }

func NewBlankAgent(name string, a account.Account) *Agent {
	return &Agent{
		Name:    name,
		Account: a,
	}
}

func (a Agent) Create() (*Agent, error) {
	id, err := createID()
	if err != nil {
		return &a, bugLog.Errorf("agent create: %+v", err)
	}
	a.UUID = id

	key, err := createKey()
	if err != nil {
		return &a, bugLog.Errorf("agent create: %+v", err)
	}
	a.Credentials.Key = key

	secret, err := createSecret()
	if err != nil {
		return &a, bugLog.Errorf("agent create: %+v", err)
	}
	a.Credentials.Secret = secret

	return &a, nil
}

func createID() (string, error) {
	id, err := generateUUID()
	if err != nil {
		return "", bugLog.Errorf("createID: %+v", err)
	}

	return id, nil
}

func createKey() (string, error) {
	key, err := generateUUID()
	if err != nil {
		return "", bugLog.Errorf("createKey: %+v", err)
	}

	return key, nil
}

func createSecret() (string, error) {
	secret, err := generateUUID()
	if err != nil {
		return "", bugLog.Errorf("generateUUID: %+v", err)
	}

	return secret, nil
}

func generateUUID() (string, error) {
	s, err := uuid.NewUUID()
	if err != nil {
		return "", bugLog.Errorf("generateUUID: %+v", err)
	}

	return s.String(), nil
}
