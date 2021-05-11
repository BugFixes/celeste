package ticketing

import (
	_ "embed"

	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v35/github"
	"github.com/mitchellh/mapstructure"
)

type Github struct {
	Client      *github.Client
	Context     context.Context
	Credentials GithubCredentials
}

type GithubRepo struct {
	Repo  string
	Owner string
}

type GithubCredentials struct {
	Credentials
	AccessToken    string
	InstallationID string
	GithubRepo
}

func NewGithub() *Github {
	return &Github{
		Context: context.Background(),
	}
}

func (g *Github) Connect() error {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 114758, 16850144, "configs/app.pem")
	if err != nil {
		return fmt.Errorf("failed to get pem file: %w", err)
	}
	g.Client = github.NewClient(&http.Client{
		Transport: itr,
	})

	return nil
}

func (g *Github) ParseCredentials(creds interface{}) error {
	type gc struct {
		AccessToken      string
		TicketingDetails struct {
			Owner          string
			Repo           string
			InstallationID string
		}
		System string
	}

	githubCreds := gc{}
	if err := mapstructure.Decode(creds, &githubCreds); err != nil {
		return fmt.Errorf("failed to decode credentials: %w", err)
	}
	g.Credentials = GithubCredentials{
		AccessToken: githubCreds.AccessToken,
		GithubRepo: GithubRepo{
			Repo:  githubCreds.TicketingDetails.Repo,
			Owner: githubCreds.TicketingDetails.Owner,
		},
		InstallationID: githubCreds.TicketingDetails.InstallationID,
	}

	return nil
}

func (g *Github) Create() error {
	title := "tester title"
	body := "tester body"

	is, resp, err := g.Client.Issues.Create(g.Context, g.Credentials.Owner, g.Credentials.Repo, &github.IssueRequest{
		Title: &title,
		Body:  &body,
	})
	if err != nil {
		return fmt.Errorf("failed to create github issue: %w", err)
	}

	fmt.Printf("Repsonse: %v\nIssue: %v\n", resp, is)

	return nil
}
