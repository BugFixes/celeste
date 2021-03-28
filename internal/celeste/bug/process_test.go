package bug_test

import (
  "errors"
  "testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/celeste/bug"
  "github.com/bugfixes/celeste/internal/config"
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
			err: errors.New("bug: parse: no body: {Resource: Path: HTTPMethod: Headers:map[tester:bob] MultiValueHeaders:map[] QueryStringParameters:map[] MultiValueQueryStringParameters:map[] PathParameters:map[] StageVariables:map[] RequestContext:{AccountID: ResourceID: OperationName: Stage: DomainName: DomainPrefix: RequestID: Protocol: Identity:{CognitoIdentityPoolID: AccountID: CognitoIdentityID: Caller: APIKey: APIKeyID: AccessKey: SourceIP: CognitoAuthenticationType: CognitoAuthenticationProvider: UserArn: UserAgent: User:} ResourcePath: Authorizer:map[] HTTPMethod: RequestTime: RequestTimeEpoch:0 APIID:} Body: IsBase64Encoded:false}"),
		},
	}

	c, err := config.BuildConfig()
	if err != nil {
	  t.Errorf("config error: %w", err)
  }

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := bug.NewProcessBug(c, *sugar).Parse(test.request)
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
			resp, err := bug.ProcessFile{}.Parse(test.request)
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
