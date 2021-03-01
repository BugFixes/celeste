package agent_test

import (
  "fmt"
  "testing"

  "github.com/bugfixes/celeste/internal/celeste/agent"
  "github.com/stretchr/testify/assert"
  "go.uber.org/zap"
)

func TestParseAgentHeaders(t *testing.T) {
  sugar := zap.NewExample().Sugar()
  defer sugar.Sync()

  tests := []struct {
    name string
    request map[string]string
    expect agent.Agent
    err error
  }{
    {
      name: "bad headers",
      request: map[string]string{},
      expect: agent.Agent{},
      err: fmt.Errorf("headers are bad"),
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      resp, err := agent.ParseAgentHeaders(test.request, sugar)
      if passed := assert.IsType(t, test.err, err); !passed {
        t.Errorf("lookup err: %w", err)
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
    name string
    request agent.Agent
    expect agent.Agent
    err error
  }{
    {
      name: "no agent",
      request: agent.Agent{
        Credentials: agent.Credentials{
          Key: "bob",
          Secret: "bill",
        },
      },
      expect: agent.Agent{},
      err: nil,
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      resp, err := test.request.LookupDetails()
      if passed := assert.IsType(t, test.err, err); !passed {
        t.Errorf("lookup err: %w", err)
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
  tests := []struct{
    name string
    request agent.Agent
    expect bool
    err error
  }{
    {
      name: "agent invalid",
      request: agent.Agent{
        Credentials: agent.Credentials{
          Key: "bob",
          Secret: "bill",
        },
      },
      expect: false,
      err: nil,
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      resp, err := test.request.ValidateID()
      if passed := assert.IsType(t, test.err, err); !passed {
        t.Errorf("validate err: %w", err)
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
