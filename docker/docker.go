package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bugfixes/celeste/internal/config"
)

func main() {
	cfg, err := config.BuildConfig()
	if err != nil {
		fmt.Printf("config error: %+v\n", err)
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.DBRegion),
		Endpoint: aws.String(cfg.AWSEndpoint),
	})
	if err != nil {
		fmt.Printf("Session Err: %+v\n", err)
		return
	}

	err = createDatabase(cfg, sess)
	if err != nil {
		fmt.Printf("create database: %+v\n", err)
		return
	}

	err = createQueue(cfg, sess)
	if err != nil {
		fmt.Printf("create queue: %+v\n", err)
		return
	}
}

// Database
func createDatabase(cfg config.Config, sess *session.Session) error {
	// Dynamo connection
	svc := dynamodb.New(sess)

	// Create bugs table
	err := createBugs(cfg, svc)
	if err != nil {
		return fmt.Errorf("create bugs: %w", err)
	}

	// Create accounts table
	err = createAccounts(cfg, svc)
	if err != nil {
		return fmt.Errorf("create accolunts: %w", err)
	}

	return nil
}

func createBugs(cfg config.Config, svc *dynamodb.DynamoDB) error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(cfg.BugsTable),
	}

	_, err := svc.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); !ok {
			return fmt.Errorf("awserr: %w", aerr)
		}
		return fmt.Errorf("unknown err: %w", err)
	}

	return nil
}

func createAccounts(cfg config.Config, svc *dynamodb.DynamoDB) error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(cfg.AccountsTable),
	}

	_, err := svc.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); !ok {
			return fmt.Errorf("awserr: %w", aerr)
		}
		return fmt.Errorf("unknown err: %w", err)
	}

	return nil
}

// Queue
func createQueue(cfg config.Config, sess *session.Session) error {
	svc := sqs.New(sess)

	input := &sqs.CreateQueueInput{
		QueueName: aws.String(cfg.QueueName),
	}
	_, err := svc.CreateQueue(input)
	if err != nil {
		return fmt.Errorf("createQueue: %w", err)
	}

	return nil
}
