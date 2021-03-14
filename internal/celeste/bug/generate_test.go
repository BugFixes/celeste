package bug_test

import (
	"testing"

	"github.com/bugfixes/celeste/internal/celeste/bug"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBug_GenerateHash(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer func() {
		_ = sugar.Sync()
	}()

	tests := []struct {
		name    string
		request bug.Bug
		expect  bug.Bug
		err     error
	}{
		{
			name: "tester hash",
			request: bug.Bug{
				Message: "tester",
			},
			expect: bug.Bug{
				Message: "tester",
				Hash:    "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.request.GenerateHash(sugar)
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

func TestBug_GenerateIdentifier(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer func() {
		_ = sugar.Sync()
	}()

	tests := []struct {
		name    string
		request bug.Bug
		expect  int
		err     error
	}{
		{
			name: "tester identifier",
			request: bug.Bug{
				Message: "tester",
			},
			expect: 36,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.request.GenerateIdentifier(sugar)
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("lookup err: %w", err)
			}
			if passed := assert.Equal(t, test.expect, len(resp.Identifier)); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, resp)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}
