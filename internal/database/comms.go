package database

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type CommsStorage struct {
	Database Database
}

type CommsCredentials struct {
	AgentID      string      `json:"agent_id"`
	CommsDetails interface{} `json:"comms_details"`
	System       string      `json:"system"`
}

func NewCommsStorage(d Database) *CommsStorage {
	return &CommsStorage{
		Database: d,
	}
}

func (c CommsStorage) FetchCredentials(agentID string) (CommsCredentials, error) {
	svc, err := c.Database.dynamoSession()
	if err != nil {
		return CommsCredentials{}, bugLog.Errorf("comms fetchCredentials session: %w", err)
	}

	filt := expression.Name("agent_id").Equal(expression.Value(agentID))
	proj := expression.NamesList(
		expression.Name("agent_id"),
		expression.Name("system"),
		expression.Name("comms_details"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		return CommsCredentials{}, bugLog.Errorf("comms fetchCredentails expression: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(c.Database.Config.CommsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		return CommsCredentials{}, bugLog.Errorf("comms fetchCredentails scan: %w", err)
	}

	ccs := []CommsCredentials{}
	if len(result.Items) == 0 {
		return CommsCredentials{}, bugLog.Errorf("comms fetchCredentails items: %w", errors.New("no items found"))
	}
	for _, i := range result.Items {
		cc := CommsCredentials{}
		if err := dynamodbattribute.UnmarshalMap(i, &cc); err != nil {
			return CommsCredentials{}, bugLog.Errorf("comms fetchCredentials unmarshal: %w", err)
		}
		ccs = append(ccs, cc)
	}

	return ccs[0], nil
}
