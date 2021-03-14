package agent

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type Credentials struct {
	Key    string
	Secret string
}

type Company struct {
	ID   string
	Name string
}

type Agent struct {
	ID   string
	Name string

	Credentials
	Company
}

func ParseAgentHeaders(headers map[string]string, logger *zap.SugaredLogger) (Agent, error) {
	a := Agent{}

	for h, v := range headers {
		l := strings.ToLower(h)
		switch l {
		case "x-agent-id":
			a.ID = v
		case "x-api-key":
			a.Key = v
		case "x-api-secret":
			a.Secret = v
		}
	}

	if (Agent{}) == a {
		logger.Errorf("headers are bad: %v", headers)
		return a, fmt.Errorf("headers are bad")
	}

	if a.ID == "" {
		return a.missingID(headers, logger)
	}

	valid, err := a.ValidateID()
	if err != nil {
		logger.Errorf("validate id failed: %v", err)
		return a, fmt.Errorf("validated id failed: %w", err)
	}
	if !valid {
		logger.Errorf("agent isn't valid: %v", a)
		return a, fmt.Errorf("agent isn't valid: %v", a)
	}

	return a, nil
}

func (a Agent) missingID(headers map[string]string, logger *zap.SugaredLogger) (Agent, error) {
	if a.Credentials.Secret == "" && a.Credentials.Key == "" {
		logger.Errorf("secret and key are blank: %v", headers)
		return a, fmt.Errorf("secret and key are blank: %v", headers)
	}

	if a.Credentials.Secret == "" {
		logger.Errorf("secret is blank: %v", headers)
		return a, fmt.Errorf("secret is blank: %v", headers)
	}

	if a.Credentials.Key == "" {
		logger.Errorf("key is blank: %v", headers)
		return a, fmt.Errorf("key is blank: %v", headers)
	}

	a, err := a.LookupDetails()
	if err != nil {
		logger.Errorf("lookup id failed: %v", err)
		return a, fmt.Errorf("lookup id failed: %w", err)
	}
	return a, nil
}

func (a Agent) ValidateID() (bool, error) {
	return false, nil
}

func (a Agent) LookupDetails() (Agent, error) {
	return Agent{}, nil
}
