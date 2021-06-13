package agent_test

import (
	"testing"

	"github.com/bugfixes/celeste/internal/celeste/account"
	"github.com/bugfixes/celeste/internal/celeste/agent"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/stretchr/testify/assert"
)

func TestParseAgentHeaders(t *testing.T) {
	tests := []struct {
		name    string
		request map[string]string
		expect  agent.Agent
		err     error
	}{
		{
			name:    "bad headers",
			request: map[string]string{},
			expect:  agent.Agent{},
			err:     bugLog.Errorf("headers are bad"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := agent.ParseAgentHeaders(test.request)
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("lookup err: %v", err)
			}
			if passed := assert.Equal(t, test.expect, resp); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, resp)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}

func TestAgent_LookupDetails(t *testing.T) {
	tests := []struct {
		name    string
		request agent.Agent
		expect  agent.Agent
		err     error
	}{
		{
			name: "no agent",
			request: agent.Agent{
				Credentials: agent.Credentials{
					Key:    "bob",
					Secret: "bill",
				},
			},
			expect: agent.Agent{},
			err:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.request.LookupDetails()
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("lookup err: %v", err)
			}
			if passed := assert.Equal(t, test.expect, resp); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, resp)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}

func TestAgent_ValidateID(t *testing.T) {
	tests := []struct {
		name    string
		request agent.Agent
		expect  bool
		err     error
	}{
		{
			name: "agent invalid",
			request: agent.Agent{
				Credentials: agent.Credentials{
					Key:    "bob",
					Secret: "bill",
				},
			},
			expect: false,
			err:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.request.ValidateID()
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("validate err: %v", err)
			}
			if passed := assert.Equal(t, test.expect, resp); !passed {
				t.Errorf("validate expect: %v, got: %v", test.expect, resp)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}

func TestAgent_Create(t *testing.T) {
	tests := []struct {
		name    string
		request *agent.Agent
		expect  *agent.Agent
		err     error
	}{
		{
			name:    "create",
			request: agent.NewAgent("tester", account.Account{}),
			expect:  &agent.Agent{},
			err:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.request.Create()
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("create err: %v", err)
			}
			if passed := assert.IsType(t, test.expect, resp); !passed {
				t.Errorf("create expect: %v", resp)
			}
		})
	}
}
