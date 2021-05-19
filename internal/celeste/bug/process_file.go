package bug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/comms"
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

func (p ProcessFile) GenerateBugInfo(bug *Bug, agentID string) error {
	bug.Agent.ID = agentID
	if err := bug.GenerateHash(&p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed hash: %+v", err)
		return fmt.Errorf("generate bug info failed hash: %w", err)
	}
	if err := bug.GenerateIdentifier(&p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed identifier: %+v", err)
		return fmt.Errorf("generate bug info failed identifier: %w", err)
	}
	if err := bug.ReportedTimes(p.Config, &p.Logger); err != nil {
		p.Logger.Errorf("generate bug info failed reportedTimes: %+v", err)
		return fmt.Errorf("generate bug info failed reportedTimes: %w", err)
	}

	bug.LevelNumber = ConvertLevelFromString(bug.Level, &p.Logger)
	bug.Posted = time.Now()

	return nil
}

func (p ProcessFile) GenerateTicket(bug *Bug) error {
	ticket := ticketing.Ticket{
		Level:         bug.Level,
		LevelNumber:   fmt.Sprint(bug.LevelNumber),
		Bug:           bug.Bug,
		Raw:           bug.Raw,
		AgentID:       bug.Agent.ID,
		Line:          bug.Line,
		File:          bug.File,
		TimesReported: bug.TimesReported,
	}

	if err := ticketing.NewTicketing(p.Config, p.Logger).CreateTicket(&ticket); err != nil {
		return fmt.Errorf("generate ticket failed: %w", err)
	}
	bug.RemoteLink = ticket.RemoteLink
	bug.TicketSystem = ticket.RemoteSystem

	return nil
}

func (p ProcessFile) GenerateComms(bug *Bug) error {
	if err := comms.NewComms(p.Config, p.Logger).SendComms(comms.CommsPackage{
		AgentID:      bug.Agent.ID,
		Message:      "tester message",
		Link:         bug.RemoteLink,
		TicketSystem: bug.TicketSystem,
	}); err != nil {
		return fmt.Errorf("file generateComms: %w", err)
	}

	return nil
}

func (p ProcessFile) errorReport(w http.ResponseWriter, textError string, wrappedError error) {
	p.Logger.Errorf("processFile errorReport: %+v", wrappedError)
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(struct {
		Error     string
		FullError string
	}{
		Error:     textError,
		FullError: fmt.Sprintf("%+v", wrappedError),
	}); err != nil {
		p.Logger.Errorf("processFile errorReport json: %+v", err)
	}
}

func (p ProcessFile) FileBugHandler(w http.ResponseWriter, r *http.Request) {
	bug := Bug{}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
		p.errorReport(w, "failed to decode bug", err)
		return
	}

	if err := p.GenerateBugInfo(&bug, r.Header.Get("X-API-KEY")); err != nil {
		p.errorReport(w, "failed to generate info", err)
		return
	}

	if err := p.GenerateTicket(&bug); err != nil {
		p.errorReport(w, "failed to generate ticket", err)
		return
	}

	if err := p.GenerateComms(&bug); err != nil {
		p.errorReport(w, "failed to generate comms", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
