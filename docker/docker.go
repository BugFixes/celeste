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
	_, err := svc.CreateTable(&dynamodb.CreateTableInput{
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
  })
	if err != nil {
		if aerr, ok := err.(awserr.Error); !ok {
			return fmt.Errorf("awserr: %w", aerr)
		}
		return fmt.Errorf("unknown err: %w", err)
	}

	return nil
}

func createAccounts(cfg config.Config, svc *dynamodb.DynamoDB) error {
	_, err := svc.CreateTable(&dynamodb.CreateTableInput{
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
  })
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

	_, err := svc.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String(cfg.QueueName),
  })
	if err != nil {
		return fmt.Errorf("createQueue: %w", err)
	}

	err = injectQueueItem(cfg, svc)
	if err != nil {
	  return fmt.Errorf("createQueue: %w", err)
  }

	return nil
}

func injectQueueItem(cfg config.Config, svc *sqs.SQS) error {
  result, err := svc.SendMessage(&sqs.SendMessageInput{
    MessageAttributes: map[string]*sqs.MessageAttributeValue{
      "clientId": &sqs.MessageAttributeValue{
        DataType: aws.String("String"),
        StringValue: aws.String("testClient"),
      },
    },
    MessageBody: aws.String("testMessage"),
    QueueUrl: aws.String(fmt.Sprintf("%s/queue/%s", cfg.AWSEndpoint, cfg.QueueName)),
  })
  if err != nil {
    return fmt.Errorf("injectQueue: %w", err)
  }

  fmt.Printf("result: %s\n", result)
  return nil
}
