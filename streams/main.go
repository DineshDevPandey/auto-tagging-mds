package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/auto-tagging-mds/database"
	"github.com/auto-tagging-mds/database/dynamodb"

	"github.com/auto-tagging-mds/database/models"
	"github.com/auto-tagging-mds/utils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type streamSvc struct {
	db            database.Database
	tableName     models.Tables
	dbCallTimeout time.Duration
	logLevel      string
}

func initSvc() (*streamSvc, error) {
	fmt.Printf("stream started : initSvc\n")
	tablesName := utils.InitTablesName()

	var db database.Database
	db, err := dynamodb.New(tablesName)
	if err != nil {
		fmt.Printf("dynamodb connection error : %v\n", err)
		return nil, err
	}

	return &streamSvc{
		db:            db,
		dbCallTimeout: 2 * time.Second,
	}, nil
}

func (sr *streamSvc) streamHandler(ctx context.Context, event models.DynamoDBEvent) error {

	fmt.Printf(" %v stream started : streamHandler\n", strings.Repeat("*", 30))

	rules, err := sr.db.GetAllRules()
	if err != nil {
		return nil
	}

	services, err := sr.db.GetAllServices()
	if err != nil {
		return nil
	}

	fmt.Printf("rule count : %v service count : %v\n", len(rules), len(services))

	for ii, record := range event.Records {

		fmt.Printf("stream : current %v total : %v\n", ii, len(event.Records))

		change := record.Change
		newImage := change.NewImage
		oldImage := change.OldImage

		var oldData models.StreamData
		var newData models.StreamData

		err := dynamodbattribute.UnmarshalMap(newImage, &newData)
		if err != nil {
			fmt.Println("UnmarshalMap error :", err)
			return err
		}

		err = dynamodbattribute.UnmarshalMap(oldImage, &oldData)
		if err != nil {
			fmt.Println("UnmarshalMap error :", err)
			return err
		}

		entity := utils.GetEntityType(newData.PK)
		switch record.EventName {
		case "MODIFY":
			switch entity {
			case utils.SERVICE:
				// do tag analysis
				if oldData.Description != newData.Description {
					// fetch rules and add tags in service
					err := sr.db.AttachTagWithService(newData, rules)
					if err != nil {
						return err
					}
				}
			case utils.RULE:
				// do tag aanalysis
				// may need to update services (tag analysys)
				fmt.Println("Rule modified")
				err := sr.db.ProcessRuleForServices(newData, services)
				if err != nil {
					return err
				}
			case utils.TAG:
				// not in assignment scope; update services
			case utils.COMPANY:
				err := sr.db.UpdateServiceTagForSubscriberCount(newData, rules)
				if err != nil {
					return err
				}
			}
		case "INSERT":
			switch entity {
			case utils.SERVICE:
				// fetch rules and add tags in service
				fmt.Println("New services created")
				err := sr.db.AttachTagWithService(newData, rules)
				if err != nil {
					return err
				}
			case utils.RULE:
				// may need to update services (tag analysys)
				fmt.Println("New rule created")
				err := sr.db.ProcessRuleForServices(newData, services)
				if err != nil {
					return err
				}
			case utils.TAG:
				// not in assignment scope; update service
			case utils.COMPANY:
				err := sr.db.UpdateServiceTagForSubscriberCount(newData, rules)
				if err != nil {
					return err
				}
			}
		case "REMOVE":
			switch entity {
			case utils.SERVICE:
				// not in assignment scope; update companies
			case utils.RULE:
				// not in assignment scope; update services
			case utils.TAG:
				// not in assignment scope; update services
			case utils.COMPANY:
				// not in assignment scope; do nothing
			}
		}
	}
	return nil
}

func (sr *streamSvc) handler(ctx context.Context, event models.DynamoDBEvent) error {
	err := sr.streamHandler(ctx, event)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	// catch run time error
	defer utils.Recover()

	fmt.Printf("stream started : main\n")
	svc, err := initSvc()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(svc.handler)
}
