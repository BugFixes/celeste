package bug_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	bug2 "github.com/bugfixes/celeste/internal/bug"
	"github.com/stretchr/testify/assert"
)

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  bug2.Response
		err     error
	}{
		{
			name: "nothing to file",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"tester": "bob",
				},
			},
			expect: bug2.Response{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := bug2.ProcessBug{}.Parse(test.request)
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