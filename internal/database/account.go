package database

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
	Key    string
	Secret string
}

type AccountRecord struct {
	ID       string
	Name     string
	ParentID string
	Email    string
	AccountCredentials
	Level       int
	DateCreated string
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
		return fmt.Errorf("insert agent: %w", err)
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
