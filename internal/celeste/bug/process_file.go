package bug

import (
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/config"
)

type ProcessFile struct {
	Config config.Config
	Logger zap.SugaredLogger
}

func NewProcessFile(c config.Config, l zap.SugaredLogger) ProcessFile {
	return ProcessFile{
		Config: c,
		Logger: l,
	}
}

func (p ProcessFile) Name() string {
	return ""
}

func (p ProcessFile) Parse(request events.APIGatewayProxyRequest) (Response, error) {
	return Response{}, nil
}

func (p ProcessFile) Report() (Response, error) {
	return Response{}, nil
}

func (p ProcessFile) Fetch() (Response, error) {
	return Response{}, nil
}
