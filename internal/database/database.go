package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
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

func dynamoError(e error, l *zap.SugaredLogger) error {
	// nolint:errorlint
	if aerr, ok := e.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
		case dynamodb.ErrCodeResourceNotFoundException:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeResourceNotFoundException, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeResourceNotFoundException, aerr)
		case dynamodb.ErrCodeTransactionConflictException:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeTransactionConflictException, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeTransactionConflictException, aerr)
		case dynamodb.ErrCodeRequestLimitExceeded:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeRequestLimitExceeded, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeRequestLimitExceeded, aerr)
		case dynamodb.ErrCodeInternalServerError:
			l.Errorf("bug insert - %s: %w", dynamodb.ErrCodeInternalServerError, aerr)
			return bugLog.Errorf("bug insert - %s: %w", dynamodb.ErrCodeInternalServerError, aerr)
		default:
			l.Errorf("bug insert - unknown err: %w", aerr)
			return bugLog.Errorf("bug insert - unknown err: %w", aerr)
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		l.Errorf("bug insert: %w", e)
		return bugLog.Errorf("bug inster: %w", e)
	}
}
