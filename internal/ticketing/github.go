package ticketing

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	bugLog "github.com/bugfixes/go-bugfixes/logs"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v35/github"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Github struct {
	Client      *github.Client
	Context     context.Context
	Credentials GithubCredentials
	Config      config.Config
	Logger      zap.SugaredLogger
}

type GithubRepo struct {
	Repo  string `json:"repo"`
	Owner string `json:"owner"`
}

type GithubCredentials struct {
	Credentials
	AccessToken    string `json:"access_token"`
	InstallationID string `json:"installation_id"`
	GithubRepo
}

func NewGithub(c config.Config, logger zap.SugaredLogger) *Github {
	return &Github{
		Context: context.Background(),
		Config:  c,
		Logger:  logger,
	}
}

func (g *Github) Connect() error {
	installationID, err := strconv.Atoi(g.Credentials.InstallationID)
	if err != nil {
		g.Logger.Errorf("github connect installid conv: %+v", err)
		return bugLog.Errorf("github connect installid conv: %w", err)
	}

	appID, err := strconv.Atoi(os.Getenv("GITHUB_APP_ID"))
	if err != nil {
		g.Logger.Errorf("github connect appid conv: %+v", err)
		return bugLog.Errorf("github connect appid conv: %w", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(appID), int64(installationID), "configs/app.pem")
	if err != nil {
		g.Logger.Errorf("github connect keyFile: %v", err)
		return bugLog.Errorf("github connect keyFile: %w", err)
	}
	g.Client = github.NewClient(&http.Client{
		Transport: itr,
	})

	return nil
}

func (g *Github) ParseCredentials(creds interface{}) error {
	type gc struct {
		AgentID          string `json:"agent_id"`
		AccessToken      string `json:"access_token"`
		TicketingDetails struct {
			Owner          string `json:"owner"`
			Repo           string `json:"repo"`
			InstallationID string `json:"installation_id" mapstructure:"installation_id"`
		} `json:"ticketing_details"`
		System string `json:"system"`
	}

	githubCreds := gc{}
	if err := mapstructure.Decode(creds, &githubCreds); err != nil {
		g.Logger.Errorf("github parseCredentials decode: %v", err)
		return bugLog.Errorf("github parseCredenhtials decode: %w", err)
	}
	g.Credentials = GithubCredentials{
		AccessToken: githubCreds.AccessToken,
		GithubRepo: GithubRepo{
			Repo:  githubCreds.TicketingDetails.Repo,
			Owner: githubCreds.TicketingDetails.Owner,
		},
		InstallationID: githubCreds.TicketingDetails.InstallationID,
		Credentials: Credentials{
			AgentID: githubCreds.AgentID,
		},
	}

	return nil
}

func (g *Github) GenerateTemplate(ticket *Ticket) (TicketTemplate, error) {
	projectFile := ticket.File
	if strings.Index(projectFile, g.Credentials.Repo) != 0 {
		projectIndex := strings.Index(ticket.File, g.Credentials.Repo)
		if projectIndex != 0 {
			projectFile = ticket.File[(projectIndex + len(g.Credentials.Repo) + 1):]
		}
	}

	title := fmt.Sprintf("File: %s, Line: %s", projectFile, ticket.Line)
	body := fmt.Sprintf(
		"## Bug\n```\n%s\n```\n## Raw\n```\n%s\n```\n### Report number\n%d\n### Link\n[%s](../blob/main/%s#L%s)\n### Latest Report Date\n%s\n",
		ticket.Bug,
		ticket.Raw,
		ticket.TimesReported,
		projectFile,
		projectFile,
		ticket.Line,
		time.Now().Format("2006-01-02 15:04:05"))

	labels := []string{
		ticket.Level,
	}
	if ticket.TimesReported == 1 {
		labels = append(labels, firstReport)
	} else {
		labels = append(labels, multiReport)
	}

	return TicketTemplate{
		Title:  title,
		Body:   body,
		Labels: labels,
		Level:  ticket.Level,
	}, nil
}

func (g *Github) Create(ticket *Ticket) error {
	template, _ := g.GenerateTemplate(ticket)

	ticketExists, td, err := g.TicketExists(ticket)
	if err != nil {
		g.Logger.Errorf("github create TicketExists: %+v", err)
		return bugLog.Errorf("github create ticketExists: %w", err)
	}

	if ticketExists {
		return g.Update(ticket)
	}

	body := fmt.Sprintf("%s", template.Body)

	is, _, err := g.Client.Issues.Create(g.Context, g.Credentials.Owner, g.Credentials.Repo, &github.IssueRequest{
		Title:  &template.Title,
		Body:   &body,
		Labels: &template.Labels,
	})
	if err != nil {
		g.Logger.Errorf("github create githubCreate: %v", err)
		return bugLog.Errorf("github create githubCreate: %w", err)
	}
	td.RemoteID = fmt.Sprintf("%d", is.GetNumber())
	ticket.RemoteID = td.RemoteID
	ticket.RemoteLink = is.GetHTMLURL()

	if err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).StoreTicketDetails(td); err != nil {
		g.Logger.Errorf("github create store: %v", err)
		return bugLog.Errorf("github create store: %w", err)
	}

	return nil
}

func (g *Github) FetchRemoteTicket(remoteData interface{}) (Ticket, error) {
	id, err := strconv.Atoi(fmt.Sprintf("%v", remoteData))
	if err != nil {
		g.Logger.Errorf("github fetchRemoteTicket id convert: %+v", err)
		return Ticket{}, bugLog.Errorf("github fetchRemoteTicket strconv: %w", err)
	}

	is, _, err := g.Client.Issues.Get(g.Context, g.Credentials.Owner, g.Credentials.Repo, id)
	if err != nil {
		g.Logger.Errorf("github fetchRemoteTicket get: %+v", err)
		return Ticket{}, bugLog.Errorf("github fetchRemoteTicket get: %w", err)
	}

	return Ticket{
		RemoteDetails: is,
	}, nil
}

func (g *Github) Fetch(ticket *Ticket) error {
	td, err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).FindTicket(database.TicketDetails{
		AgentID: g.Credentials.AgentID,
		System:  "github",
		Hash:    GenerateHash(ticket.Raw),
	})
	if err != nil {
		g.Logger.Errorf("github fetch find: %+v", err)
		return fmt.Errorf("github fetch find: %w", err)
	}

	ticket.Hash = Hash(td.Hash)
	ticket.RemoteID = td.RemoteID
	ticket.AgentID = td.AgentID

	return nil
}

func (g *Github) Update(ticket *Ticket) error {
	err := g.Fetch(ticket)
	if err != nil {
		g.Logger.Errorf("github update fetch: %+v", err)
		return fmt.Errorf("github update fetch: %w", err)
	}

	rt, err := g.FetchRemoteTicket(ticket.RemoteID)
	if err != nil {
		g.Logger.Errorf("github update fetchRemote: %+v", err)
		return fmt.Errorf("github update fetchRemote: %w", err)
	}

	is := github.Issue{}
	if err := mapstructure.Decode(rt.RemoteDetails, &is); err != nil {
		g.Logger.Errorf("update decode: %+v", err)
		return fmt.Errorf("update decode: %w", err)
	}

	template, _ := g.GenerateTemplate(ticket)
	body := fmt.Sprintf("%s", template.Body)

	state := *is.State
	if *is.State == "closed" {
		state = "open"
	}
	es, _, err := g.Client.Issues.Edit(g.Context, g.Credentials.Owner, g.Credentials.Repo, is.GetNumber(), &github.IssueRequest{
		State: &state,
		Body:  &body,
		// Labels: &[]string{
		//   t.Level,
		//   multiReport,
		// },
	})
	if err != nil {
		g.Logger.Errorf("github update reopen: %+v", err)
		return fmt.Errorf("github update reopen: %w", err)
	}
	ticket.RemoteLink = es.GetHTMLURL()

	// if _, _, err := g.Client.Issues.Edit(g.Context, g.Credentials.Owner, g.Credentials.Repo, is.GetNumber(), &github.IssueRequest{
	// 	// Labels: &[]string{
	// 	//   t.Level,
	// 	//   multiReport,
	// 	// },
	// 	Body: &body,
	// }); err != nil {
	// 	g.Logger.Errorf("github update labels: %+v", err)
	// 	return fmt.Errorf("github update labels: %w", err)
	// }

	return nil
}

func (g *Github) TicketExists(ticket *Ticket) (bool, database.TicketDetails, error) {
	td := database.TicketDetails{
		AgentID: g.Credentials.AgentID,
		System:  "github",
		Hash:    GenerateHash(ticket.Raw),
	}
	ticketExists, err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).TicketExists(td)
	if err != nil {
		g.Logger.Errorf("github ticketExists: %+v", err)
		return false, td, fmt.Errorf("github ticketExists: %w", err)
	}

	return ticketExists, td, nil
}
