package ticketing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Jira struct {
	Client      *jira.Client
	Context     context.Context
	Config      config.Config
	Logger      zap.SugaredLogger
	Credentials JiraCredentials
}

type JiraCredentials struct {
	Username    string `json:"username"`
	Token       string `json:"token"`
	Host        string `json:"host"`
	JiraProject `json:"jira_project"`
	Credentials
}

type JiraProject struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}

func NewJira(c config.Config, logger zap.SugaredLogger) *Jira {
	return &Jira{
		Context: context.Background(),
		Config:  c,
		Logger:  logger,
	}
}

func (j *Jira) Connect() error {
	c := jira.BasicAuthTransport{
		Username: j.Credentials.Username,
		Password: j.Credentials.Token,
	}

	client, err := jira.NewClient(c.Client(), j.Credentials.Host)
	if err != nil {
		j.Logger.Errorf("jira connect: %+v", err)
		return fmt.Errorf("jira connect: %w", err)
	}

	j.Client = client

	return nil
}

func (j *Jira) ParseCredentials(creds interface{}) error {
	type jc struct {
		AgentID          string `json:"agent_id"`
		AccessToken      string `json:"access_token"`
		TicketingDetails struct {
			ProjectName string `json:"project_name" mapstructure:"project_name"`
			ProjectKey  string `json:"project_key" mapstructure:"project_key"`
			Username    string `json:"username"`
			Host        string `json:"host"`
		} `json:"ticketing_details"`
		System string `json:"system"`
	}

	jiraCreds := jc{}
	if err := mapstructure.Decode(creds, &jiraCreds); err != nil {
		j.Logger.Errorf("jira parseCredentials decode: %+v", err)
		return fmt.Errorf("jira parseCredentials decode: %w", err)
	}

	j.Credentials = JiraCredentials{
		Host:     jiraCreds.TicketingDetails.Host,
		Username: jiraCreds.TicketingDetails.Username,
		Token:    jiraCreds.AccessToken,
		JiraProject: JiraProject{
			Name: jiraCreds.TicketingDetails.ProjectName,
			Key:  jiraCreds.TicketingDetails.ProjectKey,
		},
		Credentials: Credentials{
			AgentID: jiraCreds.AgentID,
		},
	}

	return nil
}

func (j *Jira) GenerateTemplate(ticket Ticket) (TicketTemplate, error) {
	projectFile := ticket.File
	title := fmt.Sprintf("File: %s, Line: %s", projectFile, ticket.Line)

	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"labels": []interface{}{
				ticket.Level,
				"first_report",
			},
			"project": map[string]interface{}{
				"key": j.Credentials.JiraProject.Key,
			},
			"issuetype": map[string]interface{}{
				"name": "Bug",
			},
			"description": map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []interface{}{
					map[string]interface{}{
						"type": "heading",
						"attrs": map[string]interface{}{
							"level": 2,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Bug",
							},
						},
					},
					map[string]interface{}{
						"type": "codeBlock",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": ticket.Bug,
							},
						},
					},
					map[string]interface{}{
						"type": "heading",
						"attrs": map[string]interface{}{
							"level": 2,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Raw",
							},
						},
					},
					map[string]interface{}{
						"type": "codeBlock",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": ticket.Raw,
							},
						},
					},
					map[string]interface{}{
						"type": "heading",
						"attrs": map[string]interface{}{
							"level": 4,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Report Number",
							},
						},
					},
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": fmt.Sprintf("%d", ticket.TimesReported),
							},
						},
					},
					map[string]interface{}{
						"type": "heading",
						"attrs": map[string]interface{}{
							"level": 4,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Latest Report Date",
							},
						},
					},
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": time.Now().Format("YYYY-MM-DD HH:mm:ss"),
							},
						},
					},
				},
			},
			"summary": title,
		},
	}

	return TicketTemplate{
		Title: title,
		Body:  body,
	}, nil
}

func (j *Jira) Create(ticket Ticket) error {
	template, _ := j.GenerateTemplate(ticket)

	exists, td, err := j.TicketExists(ticket)
	if err != nil {
		j.Logger.Errorf("jira create ticketExists: %+v", err)
		return fmt.Errorf("jira create ticketExists: %w", err)
	}
	if exists {
		return j.Update(ticket)
	}

	client := &http.Client{}
	jsond, err := json.Marshal(template.Body)
	if err != nil {
		j.Logger.Errorf("jira create marshall: %+v", err)
		return fmt.Errorf("jira create marshall: %w", err)
	}
	send := bytes.NewBuffer(jsond)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/api/3/issue", j.Credentials.Host), send)
	if err != nil {
		j.Logger.Errorf("jira create newRequest: %+v", err)
		return fmt.Errorf("jira create newRequest: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(j.Credentials.Username, j.Credentials.Token)
	resp, err := client.Do(req)
	if err != nil {
		j.Logger.Errorf("jira create do: %+v", err)
		return fmt.Errorf("jira create do: %w", err)
	}
	defer resp.Body.Close()
	readResponseBody, _ := ioutil.ReadAll(resp.Body)

	ic := jira.Issue{}
	if err := json.Unmarshal(readResponseBody, &ic); err != nil {
		j.Logger.Errorf("jira create unmarshall: %+v", err)
		return fmt.Errorf("jira create unmarshall: %w", err)
	}

	td.RemoteID = ic.ID
	if err := database.NewTicketingStorage(*database.New(j.Config, &j.Logger)).StoreTicketDetails(td); err != nil {
		j.Logger.Errorf("jira create store: %+v", err)
		return fmt.Errorf("jira create store: %w", err)
	}

	return nil
}

func (j Jira) TicketExists(ticket Ticket) (bool, database.TicketDetails, error) {
	td := database.TicketDetails{
		AgentID: j.Credentials.AgentID,
		System:  "jira",
		Hash:    GenerateHash(ticket.Raw),
	}
	ticketExists, err := database.NewTicketingStorage(*database.New(j.Config, &j.Logger)).TicketExists(td)
	if err != nil {
		j.Logger.Errorf("jira ticketExists: %+v", err)
		return false, td, fmt.Errorf("jira ticketExists: %w", err)
	}

	return ticketExists, td, nil
}

func (j Jira) Update(ticket Ticket) error {
	return nil
}

func (j Jira) FetchRemoteTicket(data interface{}) (Ticket, error) {
	return Ticket{}, nil
}

func (j Jira) Fetch(ticket Ticket) (Ticket, error) {
	return Ticket{}, nil
}
