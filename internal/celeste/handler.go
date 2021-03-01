package celeste

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/celeste/bug"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("zap failed to start: %w", err)
	}
	sugar := logger.Sugar()
	sugar.Infow("Starting Celeste")

	switch request.Path {
	case "/bug":
		sugar.Infow("bug request received")
		response, err := bug.ProcessBug(request, sugar)
		if err != nil {
			sugar.Errorf("bug request: %v, failed: %w", request, err)
			return events.APIGatewayProxyResponse{}, fmt.Errorf("process bug failed: %w\n", err)
		}
		sugar.Infow("bug request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil

	case "/file":
		sugar.Infow("file request received")
		response, err := bug.ProcessFile(request, sugar)
		if err != nil {
			sugar.Errorf("file request: %v, failed: %v", request, err)
			return events.APIGatewayProxyResponse{}, fmt.Errorf("process file failed: %w\n", err)
		}
		sugar.Infow("file request processed")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
	}

	sugar.Errorf("unknown request received: %v", request)
	return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown request received: %v", request)
}
