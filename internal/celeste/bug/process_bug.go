package bug

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/comms"
	"github.com/bugfixes/celeste/internal/logic"
	"github.com/bugfixes/celeste/internal/ticketing"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/config"
)

type ProcessBug struct {
	Config config.Config
	Logger zap.SugaredLogger

	CommsChannel string
}

func NewBug(c config.Config, l zap.SugaredLogger) ProcessBug {
	return ProcessBug{
		Config: c,
		Logger: l,
	}
}

func (p ProcessBug) Parse(request events.APIGatewayProxyRequest) (Response, error) {
	return Response{}, nil
}

func (p ProcessBug) Report() (Response, error) {

	return Response{}, nil
}

func (p ProcessBug) Fetch() (Response, error) {
	return Response{}, nil
}

func (p ProcessBug) GenerateBugInfo(bug *Bug, agentID string) error {
	bug.Agent.ID = agentID
	if err := bug.GenerateHash(&p.Logger); err != nil {
		p.Logger.Errorf("generateBugInfo generateHash: %+v", err)
		return fmt.Errorf("generateBugInfo generateHash: %w", err)
	}
	if err := bug.GenerateIdentifier(&p.Logger); err != nil {
		p.Logger.Errorf("generateBugInfo generateIdentifier: %+v", err)
		return fmt.Errorf("generateBugInfo generateIdentifier: %w", err)
	}
	if err := bug.ReportedTimes(p.Config, &p.Logger); err != nil {
		p.Logger.Errorf("generateBugInfo reportedTimes: %+v", err)
		return fmt.Errorf("generateBugInfo reportedTimes: %w", err)
	}

	bug.LevelNumber = ConvertLevelFromString(bug.Level, &p.Logger)

	return nil
}

func (p ProcessBug) GenerateTicket(bug *Bug) error {
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
		p.Logger.Errorf("generateTicket createTicket: %+v", err)
		return fmt.Errorf("generateTicket createTicket: %w", err)
	}
	bug.RemoteLink = ticket.RemoteLink
	bug.TicketSystem = ticket.RemoteSystem

	return nil
}

func (p ProcessBug) GenerateComms(bug *Bug) error {
	if err := comms.NewComms(p.Config, p.Logger).SendComms(comms.CommsPackage{
		AgentID:      bug.Agent.ID,
		Message:      "tester message",
		Link:         bug.RemoteLink,
		TicketSystem: bug.TicketSystem,
	}); err != nil {
		return fmt.Errorf("bug generateComms: %w", err)
	}

	return nil
}

func (p ProcessBug) BugHandler(w http.ResponseWriter, r *http.Request) {
	bug := Bug{}
	defer func() {
		if err := r.Body.Close(); err != nil {
			errorReport(w, p.Logger, "bugHandler body close", err)
			return
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
		errorReport(w, p.Logger, "bugHandler decode", err)
		return
	}

	if err := p.GenerateBugInfo(&bug, r.Header.Get("X-API-KEY")); err != nil {
		errorReport(w, p.Logger, "bugHandler generateBugInfo", err)
		return
	}

	if err := p.GenerateTicket(&bug); err != nil {
		errorReport(w, p.Logger, "bugHandler generateTicket", err)
		return
	}

	l := logic.NewLogic(p.Config, &p.Logger)
	if l.ShouldWeReport(logic.LogicBug{
		LastReported:  bug.LastReported,
		FirstReported: bug.FirstReported,
		TimesReported: bug.TimesReported,
	}) {
		if err := p.GenerateComms(&bug); err != nil {
			errorReport(w, p.Logger, "logHandler generateComms", err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
