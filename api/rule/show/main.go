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

type serviceShowSvc struct {
	db            database.Database
	tableName     m.Tables
	dbCallTimeout time.Duration
	logLevel      string
}

func initSvc() (*serviceShowSvc, error) {
	tablesName := u.InitTablesName()

	var db database.Database
	db, err := dynamodb.New(tablesName)
	if err != nil {
		fmt.Printf("dynamodb connection error : %v\n", err)
		return nil, err
	}

	return &serviceShowSvc{
		db:            db,
		dbCallTimeout: 2 * time.Second,
	}, nil
}

func (sc *serviceShowSvc) serviceShow(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// get query parameter
	serviceID, ok := request.QueryStringParameters["service_id"]
	if ok != true {
		return u.ApiResponse(http.StatusOK, u.EmptyStruct{})
	}

	service, err := sc.db.GetService(serviceID)
	if err != nil {
		return u.ApiResponse(http.StatusBadRequest, u.ErrorBody{
			ErrorMsg: aws.String(err.Error()),
		})
	}

	return u.ApiResponse(http.StatusOK, service)
}

func (sc *serviceShowSvc) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	events, err := sc.serviceShow(ctx, request)
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
