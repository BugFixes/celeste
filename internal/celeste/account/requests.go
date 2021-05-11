package account

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/config"
	"go.uber.org/zap"
)

type Request struct {
	Config  config.Config
	Logger  zap.SugaredLogger
	Request events.APIGatewayProxyRequest
	Account AccountCreate
}

type AccountCreate struct {
	Name        string
	Email       string
	Cellphone   int
	CountryCode int `json:"countryCode"`
}

func NewLambdaRequest(c config.Config, l zap.SugaredLogger, r events.APIGatewayProxyRequest) *Request {
	return &Request{
		Config:  c,
		Logger:  l,
		Request: r,
	}
}

func NewHTTPRequest(c config.Config, l zap.SugaredLogger) *Request {
	return &Request{
		Config: c,
		Logger: l,
	}
}

func (r Request) Parse() (Response, error) {
	// TODO Account Parse
	return Response{}, fmt.Errorf("todo: account parse")
}
