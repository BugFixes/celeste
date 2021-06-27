package ticketing

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v35/github"
	"github.com/mitchellh/mapstructure"
)

type Github struct {
	Client      *github.Client
	Context     context.Context
	Credentials GithubCredentials
	Config      config.Config
}

type GithubRepo struct {
	Repo  string `json:"repo"`
	Owner string `json:"owner"`
}

type GithubCredentials struct {
	agent.Agent
	AccessToken    string `json:"access_token"`
	InstallationID string `json:"installation_id"`
	GithubRepo
}

func NewGithub(c config.Config) *Github {
	return &Github{
		Context: context.Background(),
		Config:  c,
	}
}

func (g *Github) Connect() error {
	installationID, err := strconv.Atoi(g.Credentials.InstallationID)
	if err != nil {
		return bugLog.Errorf("github connect installid conv: %w", err)
	}

	id, err := config.GetSecret(g.Config.AWS.SecretsClient, "github_app_id")
	if err != nil {
		return bugLog.Errorf("github app id secret: %w", err)
	}
	appID, err := strconv.Atoi(id)
	if err != nil {
		return bugLog.Errorf("github connect appid conv: %w", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(appID), int64(installationID), "configs/app.pem")
	if err != nil {
		return bugLog.Errorf("github connect keyFile: %w", err)
	}
	g.Client = github.NewClient(&http.Client{
		Transport: itr,
	})

	return nil
}

func (g *Github) ParseCredentials(creds interface{}) error {
	type gc struct {
		agent.Agent
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
		return bugLog.Errorf("github parseCredenhtials decode: %w", err)
	}
	g.Credentials = GithubCredentials{
		AccessToken: githubCreds.AccessToken,
		GithubRepo: GithubRepo{
			Repo:  githubCreds.TicketingDetails.Repo,
			Owner: githubCreds.TicketingDetails.Owner,
		},
		InstallationID: githubCreds.TicketingDetails.InstallationID,
		Agent:          githubCreds.Agent,
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
		return bugLog.Errorf("github create githubCreate: %w", err)
	}
	td.RemoteID = fmt.Sprintf("%d", is.GetNumber())
	ticket.RemoteID = td.RemoteID
	ticket.RemoteLink = is.GetHTMLURL()
	td.Agent = ticket.Agent

	if err := NewTicketingStorage(g.Config).StoreTicketDetails(td); err != nil {
		return bugLog.Errorf("github create store: %w", err)
	}

	return nil
}

func (g *Github) FetchRemoteTicket(remoteData interface{}) (Ticket, error) {
	id, err := strconv.Atoi(fmt.Sprintf("%v", remoteData))
	if err != nil {
		return Ticket{}, bugLog.Errorf("github fetchRemoteTicket strconv: %w", err)
	}

	is, _, err := g.Client.Issues.Get(g.Context, g.Credentials.Owner, g.Credentials.Repo, id)
	if err != nil {
		return Ticket{}, bugLog.Errorf("github fetchRemoteTicket get: %w", err)
	}

	return Ticket{
		RemoteDetails: is,
	}, nil
}

func (g *Github) Fetch(ticket *Ticket) error {
	td, err := NewTicketingStorage(g.Config).FindTicket(TicketDetails{
		Agent:  g.Credentials.Agent,
		System: "github",
		Hash:   GenerateHash(ticket.Raw),
	})
	if err != nil {
		return bugLog.Errorf("github fetch find: %w", err)
	}

	ticket.Hash = Hash(td.Hash)
	ticket.RemoteID = td.RemoteID
	ticket.Agent = td.Agent

	return nil
}

func (g *Github) Update(ticket *Ticket) error {
	err := g.Fetch(ticket)
	if err != nil {
		return bugLog.Errorf("github update fetch: %w", err)
	}

	rt, err := g.FetchRemoteTicket(ticket.RemoteID)
	if err != nil {
		return bugLog.Errorf("github update fetchRemote: %w", err)
	}

	is := github.Issue{}
	if err := mapstructure.Decode(rt.RemoteDetails, &is); err != nil {
		return bugLog.Errorf("update decode: %w", err)
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
		return bugLog.Errorf("github update reopen: %w", err)
	}
	ticket.RemoteLink = es.GetHTMLURL()

	// if _, _, err := g.Client.Issues.Edit(g.Context, g.Credentials.Owner, g.Credentials.Repo, is.GetNumber(), &github.IssueRequest{
	// 	// Labels: &[]string{
	// 	//   t.Level,
	// 	//   multiReport,
	// 	// },
	// 	Body: &body,
	// }); err != nil {
	// 	return buglog.Errorf("github update labels: %w", err)
	// }

	return nil
}

func (g *Github) TicketExists(ticket *Ticket) (bool, TicketDetails, error) {
	td := TicketDetails{
		Agent:  g.Credentials.Agent,
		System: "github",
		Hash:   GenerateHash(ticket.Raw),
	}
	ticketExists, err := NewTicketingStorage(g.Config).TicketExists(td)
	if err != nil {
		return false, td, bugLog.Errorf("github ticketExists: %w", err)
	}

	return ticketExists, td, nil
}
