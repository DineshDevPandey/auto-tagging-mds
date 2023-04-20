package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/auto-tagging-mds/database/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

const (
	SERVICE = iota
	COMPANY
	TAG
	RULE
)

func GetPartitionKey(entity int, name string, value string) string {
	partitionKey := ""
	switch entity {
	case SERVICE:
		partitionKey = "SR#" + name
	case COMPANY:
		partitionKey = "CM#" + name
	case TAG:
		if value == "" {
			partitionKey = "TG#" + name
		} else {
			partitionKey = "TG#" + name + "#" + value
		}
	case RULE:
		partitionKey = "RL#" + name
	}
	return partitionKey
}

func GetPartitionKeyName() string {
	return "PK"
}

type EmptyStruct struct{}

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty" dynamo:"error"`
}

func ApiResponse(status int, body interface{}) (events.APIGatewayProxyResponse, error) {

	resp := events.APIGatewayProxyResponse{}
	resp.StatusCode = status
	resp.Headers = map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type,X-Amz-Date,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods": "DELETE,GET,OPTIONS,POST,PUT",
		"Content-Type":                 "application/json",
	}

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)
	return resp, nil
}

func InitTablesName() models.Tables {
	t := models.Tables{}
	t.MDSTable = os.Getenv("TABLE_NAME")

	return t
}

func Recover() {
	if r := recover(); r != nil {
		fmt.Printf("Recovering from panic in handler error is: %v \n", r)
	}
}

func GetUUID() string {
	uuid := fmt.Sprintf("%s", uuid.New())
	return uuid
}

func DateString(t string) string {
	//set timezone,
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Now().Format("2006-01-02 15:04:05")
	}

	if t == "datetime" {
		return time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	return time.Now().In(loc).Format("2006-01-02 15:04:05")
}
