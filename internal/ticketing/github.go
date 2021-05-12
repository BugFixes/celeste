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

func (g *Github) Connect() error {
	installationId, err := strconv.Atoi(g.Credentials.InstallationID)
	if err != nil {
		return fmt.Errorf("github connect convert installation id: %w", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 114758, int64(installationId), "configs/app.pem")
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

func (g *Github) Create(ticket Ticket) error {
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
		labels = append(labels, "first report")
	} else {
		labels = append(labels, "multiple reports")
	}

	is, _, err := g.Client.Issues.Create(g.Context, g.Credentials.Owner, g.Credentials.Repo, &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	})
	if err != nil {
		g.Logger.Errorf("github create ticket failed create: %v", err)
		return fmt.Errorf("github create ticket failed create: %w", err)
	}

	td := database.TicketDetails{
		AgentID:  g.Credentials.Credentials.AgentID,
		RemoteID: strconv.FormatInt(*is.ID, 10),
		System:   "github",
	}

	if err := database.NewTicketingStorage(*database.New(g.Config, &g.Logger)).StoreTicketDetails(td); err != nil {
		g.Logger.Errorf("github create ticket failed store: %v", err)
		return fmt.Errorf("github create ticket failed store: %w", err)
	}

	return nil
}
