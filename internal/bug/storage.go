package bug

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type BugStorage struct {
	Config config.Config
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

func NewBugStorage(c config.Config) *BugStorage {
	return &BugStorage{
		Config: c,
	}
}

func dynamoError(e error) error {
	// nolint:errorlint
	if aerr, ok := e.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
		case dynamodb.ErrCodeResourceNotFoundException:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeResourceNotFoundException, aerr)
		case dynamodb.ErrCodeTransactionConflictException:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeTransactionConflictException, aerr)
		case dynamodb.ErrCodeRequestLimitExceeded:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeRequestLimitExceeded, aerr)
		case dynamodb.ErrCodeInternalServerError:
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeInternalServerError, aerr)
		default:
			return bugLog.Errorf("bug insert - unknown err: %w", aerr)
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		return bugLog.Errorf("bug inster: %w", e)
	}
}

func (b BugStorage) dynamoSession() (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(b.Config.DBRegion),
		Endpoint: aws.String(b.Config.AWSEndpoint),
	})
	if err != nil {
		return nil, bugLog.Errorf("session: %w", err)
	}

	return dynamodb.New(sess), nil
}

func (b BugStorage) Insert(data BugRecord) error {
	svc, err := b.dynamoSession()
	if err != nil {
		return bugLog.Errorf("insert bug dynamo session failed: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return bugLog.Errorf("insert bug marshal failed: %w", err)
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(b.Config.BugsTable),
	})
	if err != nil {
		return dynamoError(err)
	}

	return nil
}

func (b BugStorage) FindAndStore(data BugRecord) (BugRecord, error) {
	bugRecords, err := b.Find(data)
	if err != nil {
		return BugRecord{}, bugLog.Errorf("bugstorage findAndStore find: %w", err)
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

	svc, err := b.dynamoSession()
	if err != nil {
		return brs, bugLog.Errorf("bug findAndStore session: %w", err)
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
		return brs, bugLog.Errorf("bug findAndStore build: %w", err)
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(b.Config.BugsTable),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	if err != nil {
		return brs, bugLog.Errorf("bug findAndStore scan: %w", err)
	}

	if len(result.Items) == 0 {
		return brs, nil
	}

	for _, i := range result.Items {
		bri := BugRecord{}
		if err := dynamodbattribute.UnmarshalMap(i, &bri); err != nil {
			return brs, bugLog.Errorf("bug findAndStore unmarshall: %w", err)
		}

		trn, err := strconv.Atoi(bri.TimesReported)
		if err != nil {
			return brs, bugLog.Errorf("bug findAndStore atoi: %w", err)
		}

		lr, err := time.Parse(DateFormat, bri.LastReported)
		if err != nil {
			return brs, bugLog.Errorf("bug findAndStore lastReportedParse: %w", err)
		}
		bri.LastReportedTime = lr

		fr, err := time.Parse(DateFormat, bri.FirstReported)
		if err != nil {
			return brs, bugLog.Errorf("bug findAndStore firstReportedParse: %w", err)
		}
		bri.FirstReportedTime = fr

		bri.TimesReportedNumber = trn + 1
		brs = append(brs, bri)
	}

	return brs, nil
}

func (b BugStorage) Store(data BugRecord) error {
	svc, err := b.dynamoSession()
	if err != nil {
		return bugLog.Errorf("bug store dynamosession failed: %w", err)
	}

	data.FirstReported = time.Now().Format(DateFormat)
	data.LastReported = time.Now().Format(DateFormat)

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return bugLog.Errorf("bug store marshal failed: %w", err)
	}

	if _, err := svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(b.Config.BugsTable),
	}); err != nil {
		return bugLog.Errorf("bug store putitem failed: %w", err)
	}

	return nil
}

func (b BugStorage) Update(data BugRecord) error {
	svc, err := b.dynamoSession()
	if err != nil {
		return bugLog.Errorf("bug update dynamosession failed: %w", err)
	}

	data.LastReported = time.Now().Format(DateFormat)

	if _, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tr": {
				S: aws.String(fmt.Sprint(data.TimesReportedNumber)),
			},
		},
		TableName: aws.String(b.Config.BugsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(data.ID),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set times_reported = :tr"),
	}); err != nil {
		return bugLog.Errorf("bug update updateItem failed: %w", err)
	}

	return nil
}
