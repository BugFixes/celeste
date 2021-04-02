package database

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type BugStorage struct {
	Database Database
}

type BugRecord struct {
	ID    string
	Agent string
	Level int
	Hash  string
	Full  interface{}
}

func NewBugStorage(d Database) *BugStorage {
	return &BugStorage{
		Database: d,
	}
}

func (b BugStorage) Insert(data BugRecord) error {
	svc, err := b.Database.dynamoSession()
	if err != nil {
		b.Database.Logger.Errorf("insert bug: %w", err)
		return fmt.Errorf("insert bug: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(data.ID),
			},
			"hash": {
				S: aws.String(data.Hash),
			},
			"agent": {
				S: aws.String(data.Agent),
			},
			"level": {
				N: aws.String(strconv.Itoa(data.Level)),
			},
			"full": {
				S: aws.String(data.Full.(string)),
			},
		},
		TableName: aws.String(b.Database.Config.BugsTable),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return dynamoError(err, b.Database.Logger)
	}

	return nil
}
