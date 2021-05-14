package ticketing

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"

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

const (
	firstReport = "first report"
	multiReport = "multiple reports"
)

func (g *Github) Connect() error {
	installationID, err := strconv.Atoi(g.Credentials.InstallationID)
	if err != nil {
		g.Logger.Errorf("github connet convert installation id: %+v", err)
		return fmt.Errorf("github connect convert installation id: %w", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 114758, int64(installationID), "configs/app.pem")
	if err != nil {
		g.Logger.Errorf("github failed to get pem file: %v", err)
		return fmt.Errorf("github failed to get pem file: %w", err)
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
		g.Logger.Errorf("github failed to decode credentials: %v", err)
		return fmt.Errorf("github failed to decode credentials: %w", err)
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

func (g *Github) GenerateTemplate(ticket Ticket) (TicketTemplate, error) {
	projectFile := ticket.File
	if strings.Index(projectFile, g.Credentials.Repo) != 0 {
		projectIndex := strings.Index(ticket.File, g.Credentials.Repo)
		if projectIndex != 0 {
			projectFile = ticket.File[(projectIndex + len(g.Credentials.Repo) + 1):]
		}
	}

	title := fmt.Sprintf("File: %s, Line: %s", projectFile, ticket.Line)
	body := fmt.Sprintf(
		"## Bug\n```\n%s\n```\n## Raw\n```\n%s\n```\n### Report number\n%d\n### Link\n[%s](../blob/main/%s#L%s)",
		ticket.Bug,
		ticket.Raw,
		ticket.TimesReported,
		projectFile,
		projectFile,
		ticket.Line)

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

func (g *Github) Create(ticket Ticket) error {
	template, _ := g.GenerateTemplate(ticket)

	td := database.TicketDetails{
		AgentID: g.Credentials.Credentials.AgentID,
		System:  "github",
		Hash:    GenerateHash(ticket.Raw),
	}
	ticketExists, err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).TicketExists(td)
	if err != nil {
		g.Logger.Errorf("github create ticketExists: %+v", err)
		return fmt.Errorf("github create ticketExists: %w", err)
	}

	if ticketExists {
		return g.Update(ticket)
	}

	is, _, err := g.Client.Issues.Create(g.Context, g.Credentials.Owner, g.Credentials.Repo, &github.IssueRequest{
		Title:  &template.Title,
		Body:   &template.Body,
		Labels: &template.Labels,
	})
	if err != nil {
		g.Logger.Errorf("github create ticket failed create: %v", err)
		return fmt.Errorf("github create ticket failed create: %w", err)
	}
	td.RemoteID = fmt.Sprintf("%d", is.GetNumber())

	if err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).StoreTicketDetails(td); err != nil {
		g.Logger.Errorf("github create ticket failed store: %v", err)
		return fmt.Errorf("github create ticket failed store: %w", err)
	}

	return nil
}

func (g *Github) FetchRemoteTicket(remoteData interface{}) (Ticket, error) {
	id, err := strconv.Atoi(fmt.Sprintf("%v", remoteData))
	if err != nil {
		g.Logger.Errorf("github fetchRemoteTicket id convert: %+v", err)
		return Ticket{}, fmt.Errorf("github fetchRemoteTicket id convert: %w", err)
	}

	is, _, err := g.Client.Issues.Get(g.Context, g.Credentials.Owner, g.Credentials.Repo, id)
	if err != nil {
		g.Logger.Errorf("github fetchRemoteTicket get: %+v", err)
		return Ticket{}, fmt.Errorf("github fetchRemoteTicket get: %w", err)
	}

	return Ticket{
		RemoteDetails: is,
	}, nil
}

func (g *Github) Fetch(ticket Ticket) (Ticket, error) {
	td, err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).FindTicket(database.TicketDetails{
		AgentID: g.Credentials.Credentials.AgentID,
		System:  "github",
		Hash:    GenerateHash(ticket.Raw),
	})
	if err != nil {
		g.Logger.Errorf("github fetch find ticket: %+v", err)
		return Ticket{}, fmt.Errorf("github fetch find ticket: %w", err)
	}

	return Ticket{
		Hash:     Hash(td.Hash),
		RemoteID: td.RemoteID,
		AgentID:  td.AgentID,
	}, nil
}

func (g *Github) Update(ticket Ticket) error {
	t, err := g.Fetch(ticket)
	if err != nil {
		g.Logger.Errorf("github update fetch: %+v", err)
		return fmt.Errorf("github update fetch: %w", err)
	}

	rt, err := g.FetchRemoteTicket(t.RemoteID)
	if err != nil {
		g.Logger.Errorf("github update fetch remote: %+v", err)
		return fmt.Errorf("github update fetch remote: %w", err)
	}

	is := github.Issue{}
	if err := mapstructure.Decode(rt.RemoteDetails, &is); err != nil {
		g.Logger.Errorf("update decode: %+v", err)
		return fmt.Errorf("update decode: %w", err)
	}

	template, _ := g.GenerateTemplate(ticket)

	if *is.State == "closed" {
		state := "open"
		if _, _, err := g.Client.Issues.Edit(g.Context, g.Credentials.Owner, g.Credentials.Repo, is.GetNumber(), &github.IssueRequest{
			State: &state,
			Body:  &template.Body,
			// Labels: &[]string{
			//   t.Level,
			//   multiReport,
			// },
		}); err != nil {
			g.Logger.Errorf("github update reopen: %+v", err)
			return fmt.Errorf("github update reopen: %w", err)
		}
	}

	if _, _, err := g.Client.Issues.Edit(g.Context, g.Credentials.Owner, g.Credentials.Repo, is.GetNumber(), &github.IssueRequest{
		// Labels: &[]string{
		//   t.Level,
		//   multiReport,
		// },
		Body: &template.Body,
	}); err != nil {
		g.Logger.Errorf("github update labels: %+v", err)
		return fmt.Errorf("github update labels: %w", err)
	}

	return nil
}
