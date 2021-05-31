package database

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type TicketingStorage struct {
	Database Database
}

type TicketingCredentials struct {
	AgentID          string      `json:"agent_id"`
	AccessToken      string      `json:"access_token"`
	TicketingDetails interface{} `json:"ticketing_details"`
	System           string      `json:"system"`
}

func NewTicketingStorage(d Database) *TicketingStorage {
	return &TicketingStorage{
		Database: d,
	}
}

type TicketDetails struct {
	ID       string `json:"id"`
	AgentID  string `json:"agent_id"`
	RemoteID string `json:"remote_id"`
	System   string `json:"system"`
	Hash     string `json:"hash"`
}

func (t TicketingStorage) StoreCredentials(credentials TicketingCredentials) error {
	svc, err := t.Database.dynamoSession()
	if err != nil {
		t.Database.Logger.Errorf("store credentials dynamo session: %v", err)
		return bugLog.Errorf("store credentials dynamo session: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(credentials)
	if err != nil {
		t.Database.Logger.Errorf("store credentials map failed: %v", err)
		return bugLog.Errorf("store credentials map failed: %w", err)
	}

	if _, err := svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(t.Database.Config.TicketingTable),
	}); err != nil {
		t.Database.Logger.Errorf("store credentials store failed: %v", err)
		return fmt.Errorf("store credentials store failed: %w", err)
	}

	return nil
}

func (t TicketingStorage) FetchCredentials(agentID string) (TicketingCredentials, error) {
	svc, err := t.Database.dynamoSession()
	if err != nil {
		t.Database.Logger.Errorf("ticketing fetchCredentials session: %v", err)
		return TicketingCredentials{}, fmt.Errorf("ticketing fetchCredentialssession: %w", err)
	}

	filt := expression.Name("agent_id").Equal(expression.Value(agentID))
	proj := expression.NamesList(
		expression.Name("ticketing_details"),
		expression.Name("system"),
		expression.Name("access_token"),
		expression.Name("agent_id"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		t.Database.Logger.Errorf("fetch credentials failed to build expresion: %v", err)
		return TicketingCredentials{}, fmt.Errorf("fetch credentials failed to build expresion: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(t.Database.Config.TicketingTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		return TicketingCredentials{}, fmt.Errorf("ticketing failed to scan: %w", err)
	}

	tcs := []TicketingCredentials{}
	if len(result.Items) == 0 {
		return TicketingCredentials{}, fmt.Errorf("ticketing failed to find any")
	}
	for _, i := range result.Items {
		tc := TicketingCredentials{}
		if err := dynamodbattribute.UnmarshalMap(i, &tc); err != nil {
			t.Database.Logger.Errorf("failed to unmarshal details: %v", err)
			return TicketingCredentials{}, fmt.Errorf("failed to unmarshal details: %w", err)
		}
		tcs = append(tcs, tc)
	}

	return tcs[0], nil
}

func (t TicketingStorage) StoreTicketDetails(details TicketDetails) error {
	id, err := uuid.NewUUID()
	if err != nil {
		t.Database.Logger.Errorf("store ticket uuid failed: %v", err)
		return fmt.Errorf("store ticket uuid failed: %w", err)
	}
	details.ID = id.String()

	svc, err := t.Database.dynamoSession()
	if err != nil {
		t.Database.Logger.Errorf("store ticket dynamo session: %v", err)
		return fmt.Errorf("store ticket dynamo session: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(details)
	if err != nil {
		t.Database.Logger.Errorf("store ticket marshal: %v", err)
		return fmt.Errorf("store ticket marshal: %w", err)
	}

	if _, err := svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(t.Database.Config.TicketsTable),
	}); err != nil {
		t.Database.Logger.Errorf("store ticket save: %v", err)
		return fmt.Errorf("store ticket save: %w", err)
	}

	return nil
}

func (t TicketingStorage) FindTicket(details TicketDetails) (TicketDetails, error) {
	svc, err := t.Database.dynamoSession()
	if err != nil {
		t.Database.Logger.Errorf("ticketingStorage findTicket dynamosession: %+v", err)
		return TicketDetails{}, fmt.Errorf("ticketingStorage findTicket dynamosession: %w", err)
	}

	filt := expression.And(
		expression.Name("hash").Equal(expression.Value(details.Hash)),
		expression.Name("agent_id").Equal(expression.Value(details.AgentID)))
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("agent_id"),
		expression.Name("remote_id"),
		expression.Name("system"),
		expression.Name("hash"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		t.Database.Logger.Errorf("ticketStorage findTicket expression builder: %+v", err)
		return TicketDetails{}, fmt.Errorf("ticketStorage findTicket expression buider: %w", err)
	}
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(t.Database.Config.TicketsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
	})
	if err != nil {
		t.Database.Logger.Errorf("ticketingStorage findTicket scan: %+v", err)
		return TicketDetails{}, fmt.Errorf("ticketingStorage findTicket scan: %w", err)
	}

	tds := []TicketDetails{}
	if len(result.Items) == 0 {
		return TicketDetails{}, nil
	}
	for _, i := range result.Items {
		td := TicketDetails{}
		if err := dynamodbattribute.UnmarshalMap(i, &td); err != nil {
			t.Database.Logger.Errorf("ticketingStorage findTicket unmarshal: %+v", err)
			return TicketDetails{}, fmt.Errorf("ticketingStorage findTicket unmarshal: %w", err)
		}
		tds = append(tds, td)
	}

	return tds[0], nil
}

func (t TicketingStorage) TicketExists(details TicketDetails) (bool, error) {
	ticket, err := t.FindTicket(details)
	if err != nil {
		t.Database.Logger.Errorf("ticketingStorage ticketExists findTicket: %+v", err)
		return false, fmt.Errorf("ticketingStorage ticketExists findTicket: %w", err)
	}

	if ticket.ID == "" {
		return false, nil
	}

	return true, nil
}
