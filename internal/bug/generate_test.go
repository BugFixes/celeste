package bug_test

import (
	"testing"

	bug "github.com/bugfixes/celeste/internal/bug"
	"github.com/stretchr/testify/assert"
)

func TestBug_GenerateHash(t *testing.T) {
	tests := []struct {
		name    string
		request bug.Bug
		expect  bug.Bug
		err     error
	}{
		{
			name: "tester hash",
			request: bug.Bug{
				Raw: "tester",
			},
			expect: bug.Bug{
				Raw:  "tester",
				Hash: "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.GenerateHash()
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("lookup err: %+v", err)
			}
			if passed := assert.Equal(t, test.expect.Hash, test.request.Hash); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, test.request.Hash)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}

func TestBug_GenerateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		request bug.Bug
		expect  int
		err     error
	}{
		{
			name: "tester identifier",
			request: bug.Bug{
				Raw: "tester",
			},
			expect: 36,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.GenerateIdentifier()
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("lookup err: %+v", err)
			}
			if passed := assert.Equal(t, test.expect, len(test.request.Identifier)); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, test.request.Identifier)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("lookup err failed - expected: %v, got: %v", test.err, err)
			}
		})
	}
}
