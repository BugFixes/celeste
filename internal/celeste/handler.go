package celeste

import (
  "github.com/aws/aws-lambda-go/events"
  bugLog "github.com/bugfixes/go-bugfixes/logs"
  "go.uber.org/zap"

  "github.com/bugfixes/celeste/internal/celeste/account"
  "github.com/bugfixes/celeste/internal/celeste/bug"
  "github.com/bugfixes/celeste/internal/config"
)

type Celeste struct {
	Config  config.Config
	Logger  *zap.SugaredLogger
	Request events.APIGatewayProxyRequest
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Logger
	logger, err := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	if err != nil {
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("zap failed to start: %w", err)
	}
	sugar := logger.Sugar()
	sugar.Infow("Starting Celeste")

	// Config
	cfg, err := config.BuildConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("config failed to build: %w", err)
	}

	// Routes
	c := Celeste{
		Config:  cfg,
		Logger:  sugar,
		Request: request,
	}

	return c.parseLambdaRequest()
}

// nolint: gocyclo
func (c Celeste) parseLambdaRequest() (events.APIGatewayProxyResponse, error) {
	switch c.Request.Path {
	// <editor-fold desc="Bugs">
	case "/bug":
		c.Logger.Infow("bug request received")

		response, err := bug.NewBug(c.Config, *c.Logger).Parse(c.Request)
		if err != nil {
			c.Logger.Errorf("file request: %v, failed: %v", c.Request, err)
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("process file failed: %w", err)
		}
		c.Logger.Infow("bug request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
		// </editor-fold>

		// <editor-fold desc="Logs">
	case "/log":
		c.Logger.Infow("")
	// </editor-fold>

	// <editor-fold desc="Agent">
	case "/agent":
		c.Logger.Infow("agent request received")
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("todo: agent")
		// </editor-fold>

		// <editor-fold desc="Comms">
	case "/comms":
		c.Logger.Infow("comms request received")
		return events.APIGatewayProxyResponse{}, bugLog.Errorf("todo: comms")
	// </editor-fold>

	// <editor-fold desc="Account">
	case "/account":
		c.Logger.Infow("create account request received")

		response, err := account.NewLambdaRequest(c.Config, *c.Logger, c.Request).Parse()
		if err != nil {
			c.Logger.Errorf("create account request: %v", err)
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("create account request failed: %w", err)
		}
		c.Logger.Infow("create account request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Headers:    response.Headers,
			Body:       response.Body.(string),
		}, nil

	case "/account/login":
		c.Logger.Infow("account login received")
		response, err := account.NewLambdaRequest(c.Config, *c.Logger, c.Request).Login()
		if err != nil {
			c.Logger.Errorf("account login request: %v", err)
			return events.APIGatewayProxyResponse{}, bugLog.Errorf("login account request failed: %w", err)
		}
		c.Logger.Infow("login account processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body.(string),
		}, nil
		// </editor-fold>
	}

	c.Logger.Errorf("unknown request received: %v", c.Request)
	return events.APIGatewayProxyResponse{}, bugLog.Errorf("unknown request received: %v", c.Request)
}
