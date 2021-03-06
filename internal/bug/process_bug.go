package bug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/celeste/internal/comms"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/logic"
	"github.com/bugfixes/celeste/internal/ticketing"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type ProcessBug struct {
	Config config.Config

	CommsChannel string
}

func NewBug(c config.Config) ProcessBug {
	return ProcessBug{
		Config: c,
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
	bug.Agent.UUID = agentID
	if bug.Line == "" && bug.LineNumber != 0 {
		bug.Line = strconv.Itoa(bug.LineNumber)
	}

	if err := bug.GenerateHash(); err != nil {
		return bugLog.Errorf("generateBugInfo generateHash: %+v", err)
	}
	if err := bug.GenerateIdentifier(); err != nil {
		return bugLog.Errorf("generateBugInfo generateIdentifier: %+v", err)
	}
	if err := bug.ReportedTimes(p.Config); err != nil {
		return bugLog.Errorf("generateBugInfo reportedTimes: %+v", err)
	}

	bug.LevelNumber = ConvertLevelFromString(bug.Level)

	return nil
}

func (p ProcessBug) GenerateTicket(bug *Bug) error {
	ticket := ticketing.Ticket{
		Agent:         bug.Agent,
		Level:         bug.Level,
		LevelNumber:   fmt.Sprint(bug.LevelNumber),
		Bug:           bug.Bug,
		Raw:           bug.Raw,
		Line:          bug.Line,
		File:          bug.File,
		TimesReported: bug.TimesReported,
	}

	if err := ticketing.NewTicketing(p.Config).CreateTicket(&ticket); err != nil {
		return bugLog.Errorf("generateTicket createTicket: %+v", err)
	}
	bug.RemoteLink = ticket.RemoteLink
	bug.TicketSystem = ticket.RemoteSystem

	return nil
}

func (p ProcessBug) GenerateComms(bug *Bug) error {
	if err := comms.NewComms(p.Config).SendComms(comms.CommsPackage{
		Agent:        bug.Agent,
		Message:      "tester message",
		Link:         bug.RemoteLink,
		TicketSystem: bug.TicketSystem,
	}); err != nil {
		return bugLog.Errorf("bug generateComms: %+v", err)
	}

	return nil
}

func (p ProcessBug) BugHandler(w http.ResponseWriter, r *http.Request) {
	bug := Bug{}
	defer func() {
		if err := r.Body.Close(); err != nil {
			errorReport(w, "bugHandler body close", err)
			return
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
		errorReport(w, "bugHandler decode", err)
		return
	}

	if err := p.GenerateBugInfo(&bug, r.Header.Get("X-API-KEY")); err != nil {
		errorReport(w, "bugHandler generateBugInfo", err)
		return
	}

	bug.Agent.Key = r.Header.Get("X-API-KEY")
	bug.Agent.Secret = r.Header.Get("X-API-SECRET")

	if err := p.GenerateTicket(&bug); err != nil {
		errorReport(w, "bugHandler generateTicket", err)
		return
	}

	l := logic.NewLogic(p.Config)
	if l.ShouldWeReport(logic.LogicBug{
		LastReported:  bug.LastReported,
		FirstReported: bug.FirstReported,
		TimesReported: bug.TimesReported,
	}) {
		if err := p.GenerateComms(&bug); err != nil {
			errorReport(w, "logHandler generateComms", err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
