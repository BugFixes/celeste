package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type AccountStorage struct {
	Database Database
}

const (
	AccountLevelDeity = iota
	AccountLevelOwner
	AccountLevelSub
)

type AccountCredentials struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type AccountRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	Email    string `json:"email"`
	AccountCredentials
	Level       int    `json:"level"`
	DateCreated string `json:"date_created"`
}

func GetAccountLevel(level string) int {
	switch level {
	case "diety":
		return AccountLevelDeity
	case "owner":
		return AccountLevelOwner
	default:
		return AccountLevelSub
	}
}

func NewAccountStorage(d Database) *AccountStorage {
	return &AccountStorage{
		Database: d,
	}
}

func (a AccountStorage) Insert(data AccountRecord) error {
	svc, err := a.Database.dynamoSession()
	if err != nil {
		a.Database.Logger.Errorf("insert agent: %w", err)
		return bugLog.Errorf("insert agent: %w", err)
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(data.ID),
			},
			"name": {
				S: aws.String(data.Name),
			},
			"date_created": {
				S: aws.String(data.DateCreated),
			},
			"credentials": {
				M: map[string]*dynamodb.AttributeValue{
					"key": {
						S: aws.String(data.AccountCredentials.Key),
					},
					"secret": {
						S: aws.String(data.AccountCredentials.Secret),
					},
				},
			},
		},
		TableName: aws.String(a.Database.Config.AccountsTable),
	})
	if err != nil {
		return dynamoError(err, a.Database.Logger)
	}
	return nil
}

func (a AccountStorage) Fetch(id string) (AccountRecord, error) {
	return AccountRecord{}, nil
}

func (a AccountStorage) Delete(id string) error {
	return nil
}
