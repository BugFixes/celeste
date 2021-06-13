package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Database struct {
	Config config.Config
}

func New(c config.Config) *Database {
	return &Database{
		Config: c,
	}
}

func (d Database) dynamoSession() (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(d.Config.DBRegion),
		Endpoint: aws.String(d.Config.AWSEndpoint),
	})
	if err != nil {
		return nil, bugLog.Errorf("session: %w", err)
	}

	return dynamodb.New(sess), nil
}

//go:generate mockery --name=Storage
type Storage interface {
	Insert(data interface{}) error
	Fetch(data interface{}) (interface{}, error)
	Delete(data interface{}) error
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
