package main

import (
	"context"
	"fmt"
	"log"
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

	fmt.Printf("stream started : streamHandler\n")
	// fetch all rules
	// var wg sync.WaitGroup

	// wg.Add(1)
	rules, err := sr.db.GetAllRules()
	if err != nil {
		return nil
	}

	// wg.Add(1)
	// tags, err := sr.db.GetAllTags()
	// if err != nil {
	// 	return nil
	// }

	// wg.Add(1)
	// services, err := sr.db.GetAllServices()
	// if err != nil {
	// 	return nil
	// }

	// wg.Wait()

	fmt.Printf("stream started : GetAllRules : %v\n", len(rules))

	for ii, record := range event.Records {

		fmt.Printf("Service : current %v total : %v\n", ii, len(event.Records))
		continue
		fmt.Printf("Continue didn't work\n")

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
				if oldData.Description == newData.Description {
					// fetch rules and add tags in service
				}
			case utils.RULE:
				// do tag aanalysis
			case utils.TAG:
				// not in assignment scope; update services
			case utils.COMPANY:
				// not in assignment scope; do nothing
			}
		case "INSERT":
			switch entity {
			case utils.SERVICE:
				// fetch rules and add tags in service
				fmt.Println("calling AttachTagWithService")
				err := sr.db.AttachTagWithService(newData, rules)
				if err != nil {
					return err
				}
			case utils.RULE:
				// may need to update services (tag analysys)
			case utils.TAG:
				// not in assignment scope; update service
			case utils.COMPANY:
				// not in assignment scope;
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
