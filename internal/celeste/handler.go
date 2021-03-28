package celeste

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"

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
		return events.APIGatewayProxyResponse{}, fmt.Errorf("zap failed to start: %w", err)
	}
	sugar := logger.Sugar()
	sugar.Infow("Starting Celeste")

	// Config
	cfg, err := config.BuildConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("config failed to build: %w", err)
	}

	// Routes
	c := Celeste{
		Config:  cfg,
		Logger:  sugar,
		Request: request,
	}
	return c.parseRequest()
}

func (c Celeste) parseRequest() (events.APIGatewayProxyResponse, error) {
	switch c.Request.Path {
	case "/bug":
		c.Logger.Infow("bug request received")
		response, err := bug.NewProcessBug(c.Config, *c.Logger).Parse(c.Request)
		if err != nil {
			c.Logger.Errorf("bug request: %v, failed: %w", c.Request, err)
			return events.APIGatewayProxyResponse{}, fmt.Errorf("process bug failed: %w", err)
		}
		c.Logger.Infow("bug request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil

	case "/file":
		c.Logger.Infow("file request received")

		response, err := bug.NewProcessFile(c.Config, *c.Logger).Parse(c.Request)
		if err != nil {
			c.Logger.Errorf("file request: %v, failed: %v", c.Request, err)
			return events.APIGatewayProxyResponse{}, fmt.Errorf("process file failed: %w", err)
		}
		c.Logger.Infow("file request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
	}

	c.Logger.Errorf("unknown request received: %v", c.Request)
	return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown request received: %v", c.Request)
}
