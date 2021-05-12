package bug

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/config"
)

type ProcessBug struct {
	Config config.Config
	Logger zap.SugaredLogger
}

func NewBug(c config.Config, l zap.SugaredLogger) ProcessBug {
	return ProcessBug{
		Config: c,
		Logger: l,
	}
}

func (p ProcessBug) Parse(request events.APIGatewayProxyRequest) (Response, error) {
	if len(request.Body) == 0 {
		p.Logger.Errorf("bug: parse: no body: %+v", request)
		return Response{}, fmt.Errorf("bug: parse: no body: %+v", request)
	}

	switch request.HTTPMethod {
	case http.MethodPost:
	case http.MethodPut:
		return p.Report()

	case http.MethodGet:
		return p.Fetch()
	}

	return Response{}, nil
}

func (p ProcessBug) Report() (Response, error) {

	return Response{}, nil
}

func (p ProcessBug) Fetch() (Response, error) {
	return Response{}, nil
}

func (p ProcessBug) GetBugHandler(w http.ResponseWriter, r *http.Request) {
	panic("tester")
}

func (p ProcessBug) CreateBugHandler(w http.ResponseWriter, r *http.Request) {

}
