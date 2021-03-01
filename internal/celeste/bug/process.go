package bug

import (
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

func ProcessBug(request events.APIGatewayProxyRequest, logger *zap.SugaredLogger) (Response, error) {
	return Response{}, nil
}

func ProcessFile(request events.APIGatewayProxyRequest, logger *zap.SugaredLogger) (Response, error) {
	return Response{}, nil
}
