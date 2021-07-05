package handler

import (
	"github.com/aws/aws-lambda-go/events"
	account "github.com/bugfixes/celeste/internal/account"
	bug "github.com/bugfixes/celeste/internal/bug"
	bugLog "github.com/bugfixes/go-bugfixes/logs"

	"github.com/bugfixes/celeste/internal/config"
)

type Celeste struct {
	Config  config.Config
	Request events.APIGatewayProxyRequest
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bugLog.Local().Info("Starting Celeste")

	// Config
	cfg, err := config.BuildConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("config failed to build: %+v", err)
	}

	// Routes
	c := Celeste{
		Config:  cfg,
		Request: request,
	}

	return c.parseLambdaRequest()
}

// nolint: gocyclo
func (c Celeste) parseLambdaRequest() (events.APIGatewayProxyResponse, error) {
	switch c.Request.Path {
	// <editor-fold desc="Bugs">
	case "/bug":
		bugLog.Local().Info("bug request received")

		response, err := bug.NewBug(c.Config).Parse(c.Request)
		if err != nil {
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("process file failed: %+v", err)
		}
		bugLog.Local().Info("bug request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
		// </editor-fold>

		// <editor-fold desc="Logs">
	case "/log":
		bugLog.Local().Info("")
	// </editor-fold>

	// <editor-fold desc="Agent">
	case "/agent":
		bugLog.Local().Info("agent request received")
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("todo: agent")
		// </editor-fold>

		// <editor-fold desc="Comms">
	case "/comms":
		bugLog.Local().Info("comms request received")
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("todo: comms")
	// </editor-fold>

	// <editor-fold desc="Account">
	case "/account":
		bugLog.Local().Info("create account request received")

		response, err := account.NewLambdaRequest(c.Config, c.Request).Parse()
		if err != nil {
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("create account request failed: %+v", err)
		}
		bugLog.Local().Info("create account request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Headers:    response.Headers,
			Body:       response.Body.(string),
		}, nil

	case "/account/login":
		bugLog.Local().Info("account login received")
		response, err := account.NewLambdaRequest(c.Config, c.Request).Login()
		if err != nil {
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("login account request failed: %+v", err)
		}
		bugLog.Local().Info("login account processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body.(string),
		}, nil
		// </editor-fold>
	}

	bugLog.Local().Infof("unknown request received: %v", c.Request)
	return events.APIGatewayProxyResponse{}, bugLog.Errorf("unknown request received: %v", c.Request)
}
