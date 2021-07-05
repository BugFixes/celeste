package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type AgentStorage struct {
	Database Database
}

type AgentCredentials struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type AgentRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	AgentCredentials
	AccountRecord
}

func NewAgentStorage(d Database) *AgentStorage {
	return &AgentStorage{
		Database: d,
	}
}

func (a AgentStorage) Insert(data AgentRecord) error {
	svc, err := a.Database.dynamoSession()
	if err != nil {
		return bugLog.Errorf("insert agent: %+v", err)
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(data.ID),
			},
			"name": {
				S: aws.String(data.Name),
			},
			"credentials": {
				M: map[string]*dynamodb.AttributeValue{
					"key": {
						S: aws.String(data.AgentCredentials.Key),
					},
					"secret": {
						S: aws.String(data.AgentCredentials.Secret),
					},
				},
			},
			"account": {
				M: map[string]*dynamodb.AttributeValue{
					"id": {
						S: aws.String(data.AccountRecord.ID),
					},
				},
			},
		},
	})
	if err != nil {
		return dynamoError(err)
	}
	return nil
}

func (a AgentStorage) Fetch(id string) (AgentRecord, error) {
	return AgentRecord{}, nil
}

func (a AgentStorage) Delete(id string) error {
	return nil
}
