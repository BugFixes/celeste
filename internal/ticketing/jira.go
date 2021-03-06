package ticketing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/mitchellh/mapstructure"
)

type Jira struct {
	Client      *jira.Client
	Context     context.Context
	Config      config.Config
	Credentials JiraCredentials
}

type JiraCredentials struct {
	Username    string `json:"username"`
	Token       string `json:"token"`
	Host        string `json:"host"`
	JiraProject `json:"jira_project"`
	agent.Agent
}

type JiraProject struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}

func NewJira(c config.Config) *Jira {
	return &Jira{
		Context: context.Background(),
		Config:  c,
	}
}

func (j *Jira) Connect() error {
	c := jira.BasicAuthTransport{
		Username: j.Credentials.Username,
		Password: j.Credentials.Token,
	}

	client, err := jira.NewClient(c.Client(), j.Credentials.Host)
	if err != nil {
		return bugLog.Errorf("jira connect: %+v", err)
	}

	j.Client = client

	return nil
}

func (j *Jira) ParseCredentials(creds interface{}) error {
	type jc struct {
		agent.Agent
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
		return bugLog.Errorf("jira parseCredentials decode: %+v", err)
	}

	j.Credentials = JiraCredentials{
		Host:     jiraCreds.TicketingDetails.Host,
		Username: jiraCreds.TicketingDetails.Username,
		Token:    jiraCreds.AccessToken,
		JiraProject: JiraProject{
			Name: jiraCreds.TicketingDetails.ProjectName,
			Key:  jiraCreds.TicketingDetails.ProjectKey,
		},
		Agent: jiraCreds.Agent,
	}

	return nil
}

func (j Jira) generateUpdateTemplate(ticket Ticket) TicketTemplate {
	projectFile := ticket.File
	title := fmt.Sprintf("File: %s, Line: %s", projectFile, ticket.Line)
	reportLabel := strings.ReplaceAll(multiReport, " ", "_")
	oldReportLabel := strings.ReplaceAll(firstReport, " ", "_")

	body := map[string]interface{}{
		"update": map[string]interface{}{
			"labels": []interface{}{
				map[string]interface{}{
					"add": reportLabel,
				},
				map[string]interface{}{
					"remove": oldReportLabel,
				},
			},
		},
		"fields": map[string]interface{}{
			"status": map[string]interface{}{
				"name": ticket.State,
			},
			"project": map[string]interface{}{
				"key": j.Credentials.JiraProject.Key,
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
								"text": time.Now().Format("2006-01-02 15:04:05"),
							},
						},
					},
				},
			},
		},
	}

	return TicketTemplate{
		Title: title,
		Body:  body,
	}
}

func (j Jira) generateCreateTemplate(ticket Ticket) TicketTemplate {
	projectFile := ticket.File
	title := fmt.Sprintf("File: %s, Line: %s", projectFile, ticket.Line)
	reportLabel := strings.ReplaceAll(firstReport, " ", "_")
	if ticket.TimesReported > 1 {
		reportLabel = strings.ReplaceAll(multiReport, " ", "_")
	}

	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"labels": []interface{}{
				ticket.Level,
				reportLabel,
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
								"text": time.Now().Format("2006-01-02 15:04:05"),
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
	}
}

func (j *Jira) GenerateTemplate(ticket *Ticket) (TicketTemplate, error) {
	if ticket.RemoteID != "" {
		if ticket.State != "To Do" {
			return j.generateCreateTemplate(*ticket), nil
		}
		return j.generateUpdateTemplate(*ticket), nil
	}

	return j.generateCreateTemplate(*ticket), nil
}

func (j *Jira) Create(ticket *Ticket) error {
	exists, td, err := j.TicketExists(ticket)
	if err != nil {
		return bugLog.Errorf("jira create ticketExists: %+v", err)
	}
	if exists {
		return j.Update(ticket)
	}

	return j.createNew(ticket, td)
}

func (j *Jira) createNew(ticket *Ticket, td TicketDetails) error {
	template, _ := j.GenerateTemplate(ticket)

	client := &http.Client{}
	jsond, err := json.Marshal(template.Body)
	if err != nil {
		return bugLog.Errorf("jira create marshall: %+v", err)
	}
	send := bytes.NewBuffer(jsond)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/api/3/issue", j.Credentials.Host), send)
	if err != nil {
		return bugLog.Errorf("jira create newRequest: %+v", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(j.Credentials.Username, j.Credentials.Token)
	resp, err := client.Do(req)
	if err != nil {
		return bugLog.Errorf("jira create do: %+v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			bugLog.Debugf("jira createNew close response: %+v", err)
		}
	}()

	readResponseBody, _ := ioutil.ReadAll(resp.Body)

	ic := jira.Issue{}
	if err := json.Unmarshal(readResponseBody, &ic); err != nil {
		return bugLog.Errorf("jira create unmarshall: %+v", err)
	}

	td.RemoteID = ic.ID
	ticket.RemoteLink = fmt.Sprintf("%s/browse/%s", j.Credentials.Host, ic.Key)
	if err := NewTicketingStorage(j.Config).StoreTicketDetails(td); err != nil {
		return bugLog.Errorf("jira create store: %+v", err)
	}

	return nil
}

func (j Jira) TicketExists(ticket *Ticket) (bool, TicketDetails, error) {
	td := TicketDetails{
		Agent:  j.Credentials.Agent,
		System: "jira",
		Hash:   GenerateHash(ticket.Raw),
	}
	ticketExists, err := NewTicketingStorage(j.Config).TicketExists(td)
	if err != nil {
		return false, td, bugLog.Errorf("jira ticketExists: %+v", err)
	}

	return ticketExists, td, nil
}

// Update
// nolint: gocyclo
func (j Jira) Update(ticket *Ticket) error {
	err := j.Fetch(ticket)
	if err != nil {
		return bugLog.Errorf("jira update fetch: %+v", err)
	}

	rt, err := j.FetchRemoteTicket(ticket.RemoteID)
	if err != nil {
		if strings.Contains(err.Error(), "Issue does not exist") {
			td := TicketDetails{
				Agent:  j.Credentials.Agent,
				System: "jira",
				Hash:   GenerateHash(ticket.Raw),
			}
			return j.createNew(ticket, td)
		}

		return bugLog.Errorf("jira update fetchRemoteTicket: %+v", err)
	}

	rtd := jira.Issue{}
	if err := mapstructure.Decode(rt.RemoteDetails, &rtd); err != nil {
		return bugLog.Errorf("jira update decode: %+v", err)
	}

	ticket.State = rtd.Fields.Status.Name
	ticket.RemoteLink = fmt.Sprintf("%s/browse/%s", j.Credentials.Host, rtd.Key)
	switch ticket.State {
	case "Done":
		td := TicketDetails{
			Agent:  j.Credentials.Agent,
			System: "jira",
			Hash:   GenerateHash(ticket.Raw),
		}
		return j.createNew(ticket, td)
	case "In Review": // skip creating a ticket for one thats being fixed
		return nil
	}

	template, _ := j.GenerateTemplate(ticket)
	client := &http.Client{}
	jsond, err := json.Marshal(template.Body)
	if err != nil {
		return bugLog.Errorf("jira update marshall: %+v", err)
	}
	send := bytes.NewBuffer(jsond)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/rest/api/3/issue/%s", j.Credentials.Host, rtd.ID), send)
	if err != nil {
		return bugLog.Errorf("jira update newRequest: %+v", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(j.Credentials.Username, j.Credentials.Token)
	resp, err := client.Do(req)
	if err != nil {
		return bugLog.Errorf("jira update do: %+v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			bugLog.Debugf("jira update close: %+v", err)
		}
	}()

	return nil
}

func (j Jira) FetchRemoteTicket(data interface{}) (Ticket, error) {
	is, _, err := j.Client.Issue.GetWithContext(j.Context, data.(string), &jira.GetQueryOptions{})
	if err != nil {
		return Ticket{}, bugLog.Errorf("jira fetchRemoteTicket get: %+v", err)
	}

	return Ticket{
		RemoteDetails: is,
	}, nil
}

func (j Jira) Fetch(ticket *Ticket) error {
	td, err := NewTicketingStorage(j.Config).FindTicket(TicketDetails{
		Agent:  j.Credentials.Agent,
		System: "jira",
		Hash:   GenerateHash(ticket.Raw),
	})
	if err != nil {
		return bugLog.Errorf("jira fetch find: %+v", err)
	}
	ticket.Hash = Hash(td.Hash)
	ticket.RemoteID = td.RemoteID
	ticket.Agent = td.Agent

	return nil
}
