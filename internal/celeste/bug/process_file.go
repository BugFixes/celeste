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

func (p ProcessFile) GenerateBugInfo(bug Bug, agentId string) (Bug, error) {
	bug.Agent.ID = agentId
	if err := bug.GenerateHash(&p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed hash: %+v", err)
		return bug, fmt.Errorf("generate bug info failed hash: %w", err)
	}
	if err := bug.GenerateIdentifier(&p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed identifier: %+v", err)
		return bug, fmt.Errorf("generate bug info failed identifier: %w", err)
	}
	if err := bug.ReportedTimes(p.Config, &p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed reportedTimes: %+v", err)
		return bug, fmt.Errorf("generate bug info failed reportedTimes: %w", err)
	}

	bug.LevelNumber = ConvertLevelFromString(bug.Level, &p.Logger)
	bug.Posted = time.Now()

	return bug, nil
}

func (p ProcessFile) GenerateTicket(bug Bug) error {
	if err := ticketing.NewTicketing(p.Config, p.Logger).CreateTicket(ticketing.Ticket{
		Level:         bug.Level,
		LevelNumber:   fmt.Sprint(bug.LevelNumber),
		Bug:           bug.Bug,
		Raw:           bug.Raw,
		AgentID:       bug.Agent.ID,
		Line:          bug.Line,
		File:          bug.File,
		TimesReported: bug.TimesReported,
	}); err != nil {
		return fmt.Errorf("generate ticket failed: %w", err)
	}

	return nil
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

	bug, err := p.GenerateBugInfo(bug, r.Header.Get("X-API-KEY"))
	if err != nil {
		p.Logger.Errorf("bug file generate bug info failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(struct {
			Error     string
			FullError string
		}{
			Error:     "generate bug failed",
			FullError: fmt.Sprintf("%+v", err),
		}); err != nil {
			p.Logger.Errorf("bug file generate bug failed: %+v", err)
		}
		return
	}

	if err := p.GenerateTicket(bug); err != nil {
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
