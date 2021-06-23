package bug

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

//go:generate mockery --name=Process
type Process interface {
	Parse(request events.APIGatewayProxyRequest) (Response, error)
	Report() (Response, error)
	Fetch() (Response, error)
}

func errorReport(w http.ResponseWriter, textError string, wrappedError error) {
	bugLog.Debugf("processFile errorReport: %+v", wrappedError)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(struct {
		Error     string
		FullError string
	}{
		Error:     textError,
		FullError: fmt.Sprintf("%+v", wrappedError),
	}); err != nil {
		bugLog.Debugf("processFile errorReport json: %+v", err)
	}
}
