package bug

import (
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

//go:generate mockery --name=Process
type Process interface {
  Name() string

  Parse(request events.APIGatewayProxyRequest, logger *zap.SugaredLogger) (Response, error)
  Report() error
}

func ProcessBug(request events.APIGatewayProxyRequest, logger *zap.SugaredLogger) (Response, error) {
	return Response{}, nil
}

func ProcessFile(request events.APIGatewayProxyRequest, logger *zap.SugaredLogger) (Response, error) {
	return Response{}, nil
}
