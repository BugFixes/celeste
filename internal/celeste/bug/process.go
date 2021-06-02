package bug

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

//go:generate mockery --name=Process
type Process interface {
	Parse(request events.APIGatewayProxyRequest) (Response, error)
	Report() (Response, error)
	Fetch() (Response, error)
}

func errorReport(w http.ResponseWriter, l zap.SugaredLogger, textError string, wrappedError error) {
	l.Errorf("processFile errorReport: %+v", wrappedError)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(struct {
		Error     string
		FullError string
	}{
		Error:     textError,
		FullError: fmt.Sprintf("%+v", wrappedError),
	}); err != nil {
		l.Errorf("processFile errorReport json: %+v", err)
	}
}
