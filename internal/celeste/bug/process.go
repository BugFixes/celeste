package bug

import (
	"github.com/aws/aws-lambda-go/events"
)

//go:generate mockery --name=Process
type Process interface {
	Name() string

	Parse(request events.APIGatewayProxyRequest) (Response, error)
	Report() (Response, error)
	Fetch() (Response, error)
}
