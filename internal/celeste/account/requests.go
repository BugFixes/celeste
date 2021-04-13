package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/config"
	"go.uber.org/zap"
)

type Request struct {
	Config  config.Config
	Logger  zap.SugaredLogger
	Request events.APIGatewayProxyRequest
	Account interface{}
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

func (r Request) DeleteHandler(w http.ResponseWriter, hr *http.Request) {

}
func (r Request) Delete() (Response, error) {
	// TODO Account Delete
	return Response{}, fmt.Errorf("todo: account delete")
}

func (r Request) CreateHandler(w http.ResponseWriter, hr *http.Request) {
	var ac AccountCreate
	if err := json.NewDecoder(hr.Body).Decode(&ac); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode account create: %v", err), http.StatusBadRequest)
	}

	r.Account = ac

	fmt.Printf("AC: %+v", ac)
}
func (r Request) Create() (Response, error) {
	// TODO Create Account
	return Response{}, fmt.Errorf("todo: account create")
}

func (r Request) LoginHandler(w http.ResponseWriter, hr *http.Request) {

}
func (r Request) Login() (Response, error) {
	// TODO Login Account
	return Response{}, fmt.Errorf("todo: account login")
}
