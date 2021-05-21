package database

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type LogStorage struct {
	Database Database
}

type LogRecord struct {
	ID         string    `json:"id"`
	AgentID    string    `json:"agent_id"`
	Level      string    `json:"level"`
	Line       string    `json:"line"`
	File       string    `json:"file"`
	Stack      string    `json:"stack"`
	Entry      string    `json:"entry"`
	LoggedTime time.Time `json:"logged_time" dynamodbav:"-"`
	Logged     string    `json:"logged"`
}

func NewLogStorage(d Database) *LogStorage {
	return &LogStorage{
		Database: d,
	}
}

func (l LogStorage) Store(data LogRecord) error {
	svc, err := l.Database.dynamoSession()
	if err != nil {
		l.Database.Logger.Errorf("logStorage store dynamo: %+v", err)
		return fmt.Errorf("logStorage store dynamo: %+v", err)
	}

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		l.Database.Logger.Errorf("logStorage store marshal: %+v", err)
		return fmt.Errorf("logStorage store marshal: %w", err)
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(l.Database.Config.LogsTable),
	})
	if err != nil {
		l.Database.Logger.Errorf("logStorage store putItem: %+v", err)
		return fmt.Errorf("logStorage store putItem: %w", err)
	}

	return nil
}
