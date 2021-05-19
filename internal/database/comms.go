package database

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
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
		c.Database.Logger.Errorf("comms fetchCredentials session: %+v", err)
		return CommsCredentials{}, fmt.Errorf("comms fetchCredentials session: %w", err)
	}

	filt := expression.Name("agent_id").Equal(expression.Value(agentID))
	proj := expression.NamesList(
		expression.Name("agent_id"),
		expression.Name("system"),
		expression.Name("comms_details"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		c.Database.Logger.Errorf("comms fetchCredentials expression: %+v", err)
		return CommsCredentials{}, fmt.Errorf("comms fetchCredentails expression: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(c.Database.Config.CommsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		c.Database.Logger.Errorf("comms fetchCredentials scan: %+v", err)
		return CommsCredentials{}, fmt.Errorf("comms fetchCredentails scan: %w", err)
	}

	ccs := []CommsCredentials{}
	if len(result.Items) == 0 {
		c.Database.Logger.Errorf("comms fetchCredentials items: %+v", errors.New("no items found"))
		return CommsCredentials{}, fmt.Errorf("comms fetchCredentails items: %w", errors.New("no items found"))
	}
	for _, i := range result.Items {
		cc := CommsCredentials{}
		if err := dynamodbattribute.UnmarshalMap(i, &cc); err != nil {
			c.Database.Logger.Errorf("comms fetchCredentials unmarshal: %+v", err)
			return CommsCredentials{}, fmt.Errorf("comms fetchCredentials unmarshal: %w", err)
		}
		ccs = append(ccs, cc)
	}

	return ccs[0], nil
}
