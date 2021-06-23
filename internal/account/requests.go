package account

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Request struct {
	Config  config.Config
	Request events.APIGatewayProxyRequest
	Account AccountCreate
}

type AccountCreate struct {
	Name        string
	Email       string
	Cellphone   int
	CountryCode int `json:"countryCode"`
}

func NewLambdaRequest(c config.Config, r events.APIGatewayProxyRequest) *Request {
	return &Request{
		Config:  c,
		Request: r,
	}
}

func NewHTTPRequest(c config.Config) *Request {
	return &Request{
		Config: c,
	}
}

func (r Request) Parse() (Response, error) {
	// TODO Account Parse
	return Response{}, bugLog.Errorf("todo: account parse")
}
