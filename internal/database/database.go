package database

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	sess, err := config.BuildSession(d.Config)
	if err != nil {
		return nil, bugLog.Errorf("dynamoSessioN: %+v", err)
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
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeConditionalCheckFailedException, aerr)
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr)
		case dynamodb.ErrCodeResourceNotFoundException:
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeResourceNotFoundException, aerr)
		case dynamodb.ErrCodeTransactionConflictException:
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeTransactionConflictException, aerr)
		case dynamodb.ErrCodeRequestLimitExceeded:
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeRequestLimitExceeded, aerr)
		case dynamodb.ErrCodeInternalServerError:
			return bugLog.Errorf("bug insert - %s: %+v", dynamodb.ErrCodeInternalServerError, aerr)
		default:
			return bugLog.Errorf("bug insert - unknown err: %+v", aerr)
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		return bugLog.Errorf("bug inster: %+v", e)
	}
}
