package database

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/config"
)

type Database struct {
	Config config.Config
	Logger *zap.SugaredLogger
}

func New(c config.Config, l *zap.SugaredLogger) *Database {
	return &Database{
		Config: c,
		Logger: l,
	}
}

func (d Database) dynamoSession() (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(d.Config.DBRegion),
		Endpoint: aws.String(d.Config.AWSEndpoint),
	})
	if err != nil {
		d.Logger.Errorf("session: %w", err)
		return nil, fmt.Errorf("session: %w", err)
	}

	return dynamodb.New(sess), nil
}

func (d Database) InsertBug(id, hash, agent string, lvl int, full interface{}) error {
	svc, err := d.dynamoSession()
	if err != nil {
		d.Logger.Errorf("insert bug: %w", err)
		return fmt.Errorf("insert bug: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
			"hash": {
				S: aws.String(hash),
			},
			"agent": {
				S: aws.String(agent),
			},
			"level": {
				N: aws.String(strconv.Itoa(lvl)),
			},
			"full": {
				S: aws.String(full.(string)),
			},
		},
		TableName: aws.String(d.Config.BugsTable),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
			case dynamodb.ErrCodeResourceNotFoundException:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeResourceNotFoundException, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeResourceNotFoundException, aerr)
			case dynamodb.ErrCodeTransactionConflictException:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeTransactionConflictException, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeTransactionConflictException, aerr)
			case dynamodb.ErrCodeRequestLimitExceeded:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeRequestLimitExceeded, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeRequestLimitExceeded, aerr)
			case dynamodb.ErrCodeInternalServerError:
				d.Logger.Errorf("bug insert - %s: %w", dynamodb.ErrCodeInternalServerError, aerr)
				return fmt.Errorf("bug insert - %s: %w", dynamodb.ErrCodeInternalServerError, aerr)
			default:
				d.Logger.Errorf("bug insert - unknown err: %w", aerr)
				return fmt.Errorf("bug insert - unknown err: %w", aerr)
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			d.Logger.Errorf("bug insert: %w", err)
			return fmt.Errorf("bug inster: %w", err)
		}
	}

	return nil
}
