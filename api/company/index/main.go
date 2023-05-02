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

type companySvc struct {
	db            database.Database
	tableName     m.Tables
	dbCallTimeout time.Duration
	logLevel      string
}

func initSvc() (*companySvc, error) {
	tablesName := u.InitTablesName()

	var db database.Database
	db, err := dynamodb.New(tablesName)
	if err != nil {
		fmt.Printf("dynamodb connection error : %v\n", err)
		return nil, err
	}

	return &companySvc{
		db:            db,
		dbCallTimeout: 2 * time.Second,
	}, nil
}

func (sc *companySvc) serviceIndex(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	u.IsAuthorized(request)
	services, err := sc.db.GetAllCompanies()
	if err != nil {
		return u.ApiResponse(http.StatusBadRequest, u.ErrorBody{
			ErrorMsg: aws.String(err.Error()),
		})
	}

	return u.ApiResponse(http.StatusOK, services)
}

func (sc *companySvc) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	events, err := sc.serviceIndex(ctx, request)
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
