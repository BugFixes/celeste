package agent_test

import (
	"testing"

	account2 "github.com/bugfixes/celeste/internal/account"
	agent2 "github.com/bugfixes/celeste/internal/agent"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/stretchr/testify/assert"
)

func TestParseAgentHeaders(t *testing.T) {
	tests := []struct {
		name    string
		request map[string]string
		expect  agent2.Agent
		err     error
	}{
		{
			name:    "bad headers",
			request: map[string]string{},
			expect:  agent2.Agent{},
			err:     bugLog.Errorf("headers are bad"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := agent2.ParseAgentHeaders(test.request)
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
		request agent2.Agent
		expect  agent2.Agent
		err     error
	}{
		{
			name: "no agent",
			request: agent2.Agent{
				Credentials: agent2.Credentials{
					Key:    "bob",
					Secret: "bill",
				},
			},
			expect: agent2.Agent{},
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
		request agent2.Agent
		expect  bool
		err     error
	}{
		{
			name: "agent invalid",
			request: agent2.Agent{
				Credentials: agent2.Credentials{
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
		request *agent2.Agent
		expect  *agent2.Agent
		err     error
	}{
		{
			name:    "create",
			request: agent2.NewAgent("tester", account2.Account{}),
			expect:  &agent2.Agent{},
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
