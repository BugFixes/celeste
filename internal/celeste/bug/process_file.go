package bug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/ticketing"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/config"
)

type ProcessFile struct {
	Config config.Config
	Logger zap.SugaredLogger

	CommsChannel string
}

func NewFile(c config.Config, l zap.SugaredLogger) ProcessFile {
	return ProcessFile{
		Config: c,
		Logger: l,
	}
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

func (p ProcessFile) FileBugHandler(w http.ResponseWriter, r *http.Request) {
	bug := Bug{}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
		p.Logger.Errorf("bug file parse failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Body is missing",
		}); err != nil {
			p.Logger.Errorf("bug file parse failed json: %+v", err)
		}
		return
	}
	bug, err := bug.GenerateHash(&p.Logger)
	if err != nil {
		p.Logger.Errorf("bug file failed hash : %+v", err)
	}
	bug.ParsedLevel = ConvertLevelFromString(bug.Level, &p.Logger)
	bug.Posted = time.Now()

	if err := ticketing.NewTicketing(p.Config, p.Logger).CreateTicket(ticketing.Ticket{
		Level:         fmt.Sprint(bug.ParsedLevel),
		Bug:           bug.Bug,
		Raw:           bug.Raw,
		AgentID:       r.Header.Get("X-API-KEY"),
		Line:          bug.Line,
		File:          bug.File,
		ReportedTimes: 1,
	}); err != nil {
		p.Logger.Errorf("bug file ticket failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(struct {
			Error     string
			FullError string
		}{
			Error:     "Ticket failed",
			FullError: fmt.Sprintf("%+v", err),
		}); err != nil {
			p.Logger.Errorf("bug file ticket failed json: %+v", err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
