package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/auto-tagging-mds/database"

	m "github.com/auto-tagging-mds/database/models"
	u "github.com/auto-tagging-mds/utils"

	"github.com/auto-tagging-mds/database/dynamodb"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

type tagSvc struct {
	db            database.Database
	tableName     m.Tables
	dbCallTimeout time.Duration
	logLevel      string
}

func initSvc() (*tagSvc, error) {
	tablesName := u.InitTablesName()

	var db database.Database
	db, err := dynamodb.New(tablesName)
	if err != nil {
		fmt.Printf("dynamodb connection error : %v\n", err)
		return nil, err
	}

	return &tagSvc{
		db:            db,
		dbCallTimeout: 2 * time.Second,
	}, nil
}

func (sc *tagSvc) tagIndex(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u.IsAuthorized(request)
	tags, err := sc.db.GetAllTags()
	if err != nil {
		return u.ApiResponse(http.StatusBadRequest, u.ErrorBody{
			ErrorMsg: aws.String(err.Error()),
		})
	}

	return u.ApiResponse(http.StatusOK, tags)
}

func (sc *tagSvc) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	events, err := sc.tagIndex(ctx, request)
	if err != nil {
		log.Fatal(err)
	}
	return events, nil
}

func main() {
	// catch run time error
	defer u.Recover()

	svc, err := initSvc()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(svc.handler)
}
