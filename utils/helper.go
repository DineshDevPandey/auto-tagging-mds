package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/auto-tagging-mds/database/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

const (
	SERVICE = iota
	COMPANY
	TAG
	RULE
)

const (
	DESCRIPTION   = "description"
	LOCATION      = "location"
	LIKE          = "like"
	TARGETSEGMENT = "targetsegment"
)

const (
	GREATER            = "GREATER"
	LESSER             = "LESSER"
	EQUAL              = "EQUAL"
	GREATER_THAN_EQUAL = "GREATER_THAN_EQUAL"
	LESSER_THAN_EQUAL  = "LESSER_THAN_EQUAL"
)

const (
	AND = "AND"
	OR  = "OR"
)

func GetEntityType(pk string) int {
	switch pk {
	case "SR":
		return SERVICE
	case "RL":
		return RULE
	case "TG":
		return TAG
	case "CM":
		return COMPANY
	}
	return -1
}

func GetPartitionKey(entity int) string {
	partitionKey := ""
	switch entity {
	case SERVICE:
		partitionKey = "SR"
	case COMPANY:
		partitionKey = "CM"
	case TAG:
		partitionKey = "TG"
	case RULE:
		partitionKey = "RL"
	}
	return partitionKey
}

func GetRangeKey(entity int, name, value, uuid string) string {
	rangeKey := ""
	switch entity {
	case SERVICE:
		rangeKey = "SR#" + name
	case COMPANY:
		rangeKey = "CM#" + name
	case TAG:
		if value == "" {
			rangeKey = "TG#" + name
		} else {
			rangeKey = "TG#" + name + "#" + value
		}
	case RULE:
		rangeKey = "RL#" + uuid
	}
	return rangeKey
}

func GetPartitionKeyName() string {
	return "PK"
}

func GetRangeKeyName() string {
	return "SK"
}

type EmptyStruct struct{}

type MissingParameter struct {
	ErrorMsg string `json:"error,omitempty"`
}

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
	t.MDSTable = "at_mds-prod"
	// t.MDSTable = os.Getenv("TABLE_NAME")

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

func NilToEmptySlice(av map[string]*dynamodb.AttributeValue, field string) map[string]*dynamodb.AttributeValue {
	empty := []*dynamodb.AttributeValue{}
	av[field] = &dynamodb.AttributeValue{L: empty}
	return av
}

func IsTagAttachable(streamData models.StreamData, rule models.RuleResponse) bool {

	fmt.Println("inside IsTagAttachable ")

	// ruleMetadataFieldList contains MetadataField and CoRuleMetadataField
	ruleMetadataFieldList := make([]string, 0)
	// ruleKeywordList contains Keyword and CoRuleKeyword
	ruleKeywordList := make([]string, 0)

	ruleMetadataFieldList = append(ruleMetadataFieldList, rule.MetadataField)
	ruleKeywordList = append(ruleKeywordList, rule.Keyword)

	if rule.CoRuleMetadataField != "" {
		ruleMetadataFieldList = append(ruleMetadataFieldList, rule.CoRuleMetadataField)
		ruleKeywordList = append(ruleKeywordList, rule.CoRuleKeyword)
	}

	var cond [2]bool
	for i, ruleMetadataField := range ruleMetadataFieldList {
		cond[i] = matchCondition(ruleMetadataField, streamData, ruleKeywordList[i], rule.Operand, rule.RelationalOperator)
		fmt.Println("inside IsTagAttachable cond[i] :", cond[i])
	}

	switch rule.KeywordOperator {
	case AND:
		fmt.Println("inside IsTagAttachable cond[i] AND :", cond[0] && cond[1])
		return cond[0] && cond[1]
	case OR:
		fmt.Println("inside IsTagAttachable cond[i] OR :", cond[0] || cond[1])
		return cond[0] || cond[1]
	case "":
		fmt.Println("inside IsTagAttachable cond[i] blank :", cond[0])
		return cond[0]
	default:
		fmt.Println("inside IsTagAttachable default :", rule.KeywordOperator)
		return false
	}
}

func matchCondition(ruleMetadataField string, streamData models.StreamData, keyword string, operand int, relationalOperator string) bool {
	// mdValue : description value, like value
	mdValue := getMetaDataFieldValue(ruleMetadataField, streamData)
	fmt.Println("getMetaDataFieldValue : ", mdValue)

	if strings.ToLower(ruleMetadataField) == LIKE {
		likeCount, _ := strconv.Atoi(mdValue)
		switch relationalOperator {
		case GREATER:
			if likeCount > operand {
				return true
			}
		case LESSER:
			if likeCount < operand {
				return true
			}
		case GREATER_THAN_EQUAL:
			if likeCount >= operand {
				return true
			}
		case LESSER_THAN_EQUAL:
			if likeCount <= operand {
				return true
			}
		case EQUAL:
			if likeCount == operand {
				return true
			}
		}
	} else {
		keyword := strings.ToLower(keyword)
		if strings.Contains(mdValue, keyword) {
			fmt.Printf("strings.Contains : keyword %v : mdValue : %v", keyword, mdValue)
			return true
		}
	}
	return false
}

func getMetaDataFieldValue(md string, streamData models.StreamData) string {
	value := ""
	md = strings.ReplaceAll(strings.ToLower(md), " ", "")

	switch md {
	case DESCRIPTION:
		value = streamData.Description
	case LOCATION:
		value = streamData.Location
	case LIKE:
		value = fmt.Sprint(streamData.Like)
	case TARGETSEGMENT:
		value = streamData.TargetSegment
	default:
		return ""
	}

	return strings.ToLower(value)
}

func IsTagAlreadyPresent(category []models.Category, cat models.Category) bool {
	for _, tag := range category {
		if tag.Key == cat.Key && tag.Value == cat.Value {
			return true
		}
	}
	return false
}
