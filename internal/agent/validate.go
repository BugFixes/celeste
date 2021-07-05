package agent

import (
	"strings"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

func blankAgent(a, b Agent) bool {
	if a.ID != b.ID ||
		a.Secret != b.Secret ||
		a.Key != b.Key ||
		a.Name != b.Name {
		return false
	}

	return true
}

func ParseAgentHeaders(headers map[string]string) (Agent, error) {
	a := Agent{}

	for h, v := range headers {
		l := strings.ToLower(h)
		switch l {
		case "x-agent-id":
			a.UUID = v
		case "x-api-key":
			a.Key = v
		case "x-api-secret":
			a.Secret = v
		}
	}

	if blankAgent(a, Agent{}) {
		return a, bugLog.Errorf("headers are bad")
	}

	if a.UUID == "" {
		return a.missingID(headers)
	}

	valid, err := a.ValidateID()
	if err != nil {
		return a, bugLog.Errorf("validated id failed: %+v", err)
	}
	if !valid {
		return a, bugLog.Errorf("agent isn't valid: %v", a)
	}

	return a, nil
}

func (a Agent) missingID(headers map[string]string) (Agent, error) {
	if a.Credentials.Secret == "" && a.Credentials.Key == "" {
		return a, bugLog.Errorf("secret and key are blank: %v", headers)
	}

	if a.Credentials.Secret == "" {
		return a, bugLog.Errorf("secret is blank: %v", headers)
	}

	if a.Credentials.Key == "" {
		return a, bugLog.Errorf("key is blank: %v", headers)
	}

	a, err := a.LookupDetails()
	if err != nil {
		return a, bugLog.Errorf("lookup id failed: %+v", err)
	}
	return a, nil
}

func (a Agent) ValidateID() (bool, error) {
	return false, nil
}

func (a Agent) LookupDetails() (Agent, error) {
	return Agent{}, nil
}
