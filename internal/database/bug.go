package database

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type BugStorage struct {
	Database Database
}

type BugRecord struct {
	ID                  string      `json:"id"`
	AgentID             string      `json:"agent_id"`
	Level               string      `json:"level"`
	Hash                string      `json:"hash"`
	Full                interface{} `json:"full"`
	TimesReported       string      `json:"times_reported"`
	TimesReportedNumber int         `json:"times_reported_number"`
}

func NewBugStorage(d Database) *BugStorage {
	return &BugStorage{
		Database: d,
	}
}

func (b BugStorage) Insert(data BugRecord) error {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("insert bug dynamo session failed: %+v", err)
		return fmt.Errorf("insert bug dynamo session failed: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		b.Database.Logger.Errorf("insert bug marshal failed: %+v", err)
		return fmt.Errorf("insert bug marshal failed: %w", err)
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(b.Database.Config.BugsTable),
	})
	if err != nil {
		return dynamoError(err, b.Database.Logger)
	}

	return nil
}

func (b BugStorage) FindAndStore(data BugRecord) (BugRecord, error) {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore dynamo session failed: %+v", err)
		return BugRecord{}, fmt.Errorf("bug findAndStore dynamo session failed: %w", err)
	}

	filt := expression.And(
		expression.Name("hash").Equal(expression.Value(data.Hash)),
		expression.Name("agent_id").Equal(expression.Value(data.AgentID)))
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("agent_id"),
		expression.Name("level"),
		expression.Name("hash"),
		expression.Name("full"),
		expression.Name("times_reported"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore expression builder failed: %+v", err)
		return BugRecord{}, fmt.Errorf("bug findAndStore expression builder failed: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(b.Database.Config.BugsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore scan failed: %+v", err)
		return BugRecord{}, fmt.Errorf("bug findAndStore scan failed: %w", err)
	}

	brs := []BugRecord{}
	if len(result.Items) == 0 {
		data.TimesReportedNumber = 1
		return data, b.Store(data)
	}
	for _, i := range result.Items {
		bri := BugRecord{}
		if err := dynamodbattribute.UnmarshalMap(i, &bri); err != nil {
			b.Database.Logger.Errorf("bug findAndStore unmarshall failed: %+v", err)
			return BugRecord{}, fmt.Errorf("bug findAndStore unmarshall failed: %w", err)
		}

		trn, err := strconv.Atoi(bri.TimesReported)
		if err != nil {
			b.Database.Logger.Errorf("bug findAndStore convert number: %+v", err)
			return BugRecord{}, fmt.Errorf("bug findAndStore convert number: %w", err)
		}

		bri.TimesReportedNumber = trn + 1
		brs = append(brs, bri)
	}

	return brs[0], b.Update(brs[0])
}

func (b BugStorage) Store(data BugRecord) error {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("bug store dynamosession failed: %+v", err)
		return fmt.Errorf("bug store dynamosession failed: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		b.Database.Logger.Errorf("bug store marshal failed: %+v", err)
		return fmt.Errorf("bug store marshal failed: %w", err)
	}

	if _, err := svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(b.Database.Config.BugsTable),
	}); err != nil {
		b.Database.Logger.Errorf("bug store putitem failed: %+v", err)
		return fmt.Errorf("bug store putitem failed: %w", err)
	}

	return nil
}

func (b BugStorage) Update(data BugRecord) error {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("bug update dynamosession failed: %+v", err)
		return fmt.Errorf("bug update dynamosession failed: %w", err)
	}

	if _, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tr": {
				S: aws.String(fmt.Sprint(data.TimesReportedNumber + 1)),
			},
		},
		TableName: aws.String(b.Database.Config.BugsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(data.ID),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set times_reported = :tr"),
	}); err != nil {
		b.Database.Logger.Errorf("bug update updateItem failed: %+v", err)
		return fmt.Errorf("bug update updateItem failed: %w", err)
	}

	return nil
}
