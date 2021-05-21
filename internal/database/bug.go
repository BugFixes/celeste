package database

import (
	"fmt"
	"strconv"
	"time"

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
	TimesReportedNumber int         `json:"times_reported_number" dynamodbav:"-"`
	LastReportedTime    time.Time   `json:"last_reported_time" dynamodbav:"-"`
	LastReported        string      `json:"last_reported"`
	FirstReportedTime   time.Time   `json:"first_reported_time" dynamodbav:"-"`
	FirstReported       string      `json:"first_reported"`
}

const DateFormat = "2006-04-02 15:04:05"

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
	bugRecords, err := b.Find(data)
	if err != nil {
		b.Database.Logger.Errorf("bugstorage findAndStore find: %+v", err)
		return BugRecord{}, fmt.Errorf("bugstorage findAndStore find: %w", err)
	}

	if len(bugRecords) == 0 {
		data.TimesReportedNumber = 1
		data.TimesReported = "1"
		data.LastReportedTime = time.Now()
		data.LastReported = time.Now().Format(DateFormat)
		data.FirstReportedTime = time.Now()
		data.FirstReported = time.Now().Format(DateFormat)
		return data, b.Store(data)
	}

	return bugRecords[0], b.Update(bugRecords[0])
}

func (b BugStorage) Find(data BugRecord) ([]BugRecord, error) {
	brs := []BugRecord{}

	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore session: %+v", err)
		return brs, fmt.Errorf("bug findAndStore session: %w", err)
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
		expression.Name("times_reported"),
		expression.Name("last_reported"),
		expression.Name("first_reported"))
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore build: %+v", err)
		return brs, fmt.Errorf("bug findAndStore build: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(b.Database.Config.BugsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		b.Database.Logger.Errorf("bug findAndStore scan: %+v", err)
		return brs, fmt.Errorf("bug findAndStore scan: %w", err)
	}

	if len(result.Items) == 0 {
		return brs, nil
	}

	for _, i := range result.Items {
		bri := BugRecord{}
		if err := dynamodbattribute.UnmarshalMap(i, &bri); err != nil {
			b.Database.Logger.Errorf("bug findAndStore unmarshall: %+v", err)
			return brs, fmt.Errorf("bug findAndStore unmarshall: %w", err)
		}

		trn, err := strconv.Atoi(bri.TimesReported)
		if err != nil {
			b.Database.Logger.Errorf("bug findAndStore atoi: %+v", err)
			return brs, fmt.Errorf("bug findAndStore atoi: %w", err)
		}

		lr, err := time.Parse(DateFormat, bri.LastReported)
		if err != nil {
			b.Database.Logger.Errorf("bug findAndStore lastReportedParse: %+v", err)
			return brs, fmt.Errorf("bug findAndStore lastReportedParse: %w", err)
		}
		bri.LastReportedTime = lr

		fr, err := time.Parse(DateFormat, bri.FirstReported)
		if err != nil {
			b.Database.Logger.Errorf("bug findAndStore firstReportedParse: %+v", err)
			return brs, fmt.Errorf("bug findAndStore firstReportedParse: %w", err)
		}
		bri.FirstReportedTime = fr

		bri.TimesReportedNumber = trn + 1
		brs = append(brs, bri)
	}

	return brs, nil
}

func (b BugStorage) Store(data BugRecord) error {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("bug store dynamosession failed: %+v", err)
		return fmt.Errorf("bug store dynamosession failed: %w", err)
	}

	data.FirstReported = time.Now().Format(DateFormat)
	data.LastReported = time.Now().Format(DateFormat)

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

	data.LastReported = time.Now().Format(DateFormat)

	if _, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tr": {
				S: aws.String(fmt.Sprint(data.TimesReportedNumber)),
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
