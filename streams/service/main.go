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

type serviceSvc struct {
	db            database.Database
	tableName     models.Tables
	dbCallTimeout time.Duration
	logLevel      string
}

func initSvc() (*serviceSvc, error) {
	tablesName := utils.InitTablesName()

	var db database.Database
	db, err := dynamodb.New(tablesName)
	if err != nil {
		fmt.Printf("dynamodb connection error : %v\n", err)
		return nil, err
	}

	return &serviceSvc{
		db:            db,
		dbCallTimeout: 2 * time.Second,
	}, nil
}

func (sr *serviceSvc) inventory(ctx context.Context, event models.DynamoDBEvent) error {

	// fetch all rules
	rules, err := sr.db.GetAllRules()
	if err != nil {
		return nil
	}

	for ii, record := range event.Records {

		fmt.Printf("Service : current %v total : %v\n", ii, len(event.Records))
		change := record.Change
		newImage := change.NewImage 

		var oldService models.ServiceRequest
		var newService models.ServiceRequest

		err := dynamodbattribute.UnmarshalMap(newImage, &newService)
		if err != nil {
			fmt.Println("UnmarshalMap error :", err)
			return err
		}

		switch record.EventName {
		case "MODIFY":
			err := dynamodbattribute.UnmarshalMap(newImage, &oldService)
			if err != nil {
				fmt.Println("UnmarshalMap error :", err)
				return err
			}
			if oldService.Description == newService.Description {
				return nil
			}
			fallthrough

		case "INSERT":
			err := sr.db.AttachTagWithService(newService, rules)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sr *serviceSvc) handler(ctx context.Context, event models.DynamoDBEvent) error {
	err := sr.inventory(ctx, event)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	// catch run time error
	defer utils.Recover()

	svc, err := initSvc()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(svc.handler)
}
