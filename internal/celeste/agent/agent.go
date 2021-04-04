package agent

import (
	"fmt"

	"github.com/bugfixes/celeste/internal/celeste/account"
	"github.com/google/uuid"
)

type Credentials struct {
	Key    string
	Secret string
}

type Agent struct {
	ID   string
	Name string

	Credentials
	account.Account
}

//go:generate mockery --name=Agents
type Agents interface {
	Create() (*Agent, error)
	Delete(a Agent) error
}

func NewAgent(name string, account account.Account) *Agent {
	return &Agent{
		Name:    name,
		Account: account,
	}
}

func (a Agent) Create() (*Agent, error) {
	id, err := createID()
	if err != nil {
		return &a, fmt.Errorf("agent create: %w", err)
	}
	a.ID = id

	key, err := createKey()
	if err != nil {
		return &a, fmt.Errorf("agent create: %w", err)
	}
	a.Credentials.Key = key

	secret, err := createSecret()
	if err != nil {
		return &a, fmt.Errorf("agent create: %w", err)
	}
	a.Credentials.Secret = secret

	return &a, nil
}

func createID() (string, error) {
	id, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("createID: %w", err)
	}

	return id, nil
}

func createKey() (string, error) {
	key, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("createKey: %w", err)
	}

	return key, nil
}

func createSecret() (string, error) {
	secret, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("generateUUID: %w", err)
	}

	return secret, nil
}

func generateUUID() (string, error) {
	s, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("generateUUID: %w", err)
	}

	return s.String(), nil
}
