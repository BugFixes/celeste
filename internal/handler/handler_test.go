package handler_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestAccount(t *testing.T) {
	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  events.APIGatewayProxyResponse
		err     error
	}{
		{
			name: "test1",
			request: events.APIGatewayProxyRequest{
				Path:       "/account",
				HTTPMethod: "POST",
				Body:       `{"name":"tester","email":"tester@test.com"}`,
			},
			expect: events.APIGatewayProxyResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := handler.Handler(test.request)
			if passed := assert.IsType(t, test.err, err); !passed {
				t.Errorf("Account err type: %v, %+v", err, test.err)
			}
			if passed := assert.Equal(t, test.err, err); !passed {
				t.Errorf("Account err equal: %v, %+v", err, test.err)
			}
			if passed := assert.Equal(t, test.expect, resp); !passed {
				t.Errorf("Account equal: %v, %+v", resp, test.expect)
			}
		})
	}
}
