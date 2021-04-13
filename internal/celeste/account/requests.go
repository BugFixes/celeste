package account

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/config"
	"go.uber.org/zap"
)

type AccountRequest struct {
	Config  config.Config
	Logger  zap.SugaredLogger
	Request events.APIGatewayProxyRequest
}

func NewAccountRequest(c config.Config, l zap.SugaredLogger, r events.APIGatewayProxyRequest) *AccountRequest {
	return &AccountRequest{
		Config:  c,
		Logger:  l,
		Request: r,
	}
}

func (ar AccountRequest) Create() (Response, error) {
	// TODO Create Account
	return Response{}, nil
}

func (ar AccountRequest) Login() (Response, error) {
	// TODO Login Account
	return Response{}, nil
}
