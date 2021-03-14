package bug_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/celeste/bug"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProcessBug(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer func() {
		_ = sugar.Sync()
	}()

	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  bug.Response
		err     error
	}{
		{
			name: "no bug",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"tester": "bob",
				},
			},
			expect: bug.Response{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := bug.ProcessBug(test.request, sugar)
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

func TestProcessFile(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer func() {
		_ = sugar.Sync()
	}()

	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  bug.Response
		err     error
	}{
		{
			name: "nothing to file",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"tester": "bob",
				},
			},
			expect: bug.Response{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := bug.ProcessFile(test.request, sugar)
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
