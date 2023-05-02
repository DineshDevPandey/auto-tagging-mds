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

const DELIMITER = "*=*"

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
	TARGETSEGMENT = "target_segment"
	PRICING       = "pricing"
	BUSINESSMODEL = "business_model"
	DEPLOYMENT    = "deployment"
	STAGE         = "stage"
)

const (
	GREATER_THAN       = "GREATER_THAN"
	LESSER_THAN        = "LESSER_THAN"
	EQUAL              = "EQUAL"
	GREATER_THAN_EQUAL = "GREATER_THAN_EQUAL"
	LESSER_THAN_EQUAL  = "LESSER_THAN_EQUAL"
)

const (
	CONTAIN            = "CONTAIN"
	RELATION           = "RELATION"
	SUBSCRIPTION_COUNT = "SUBSCRIPTION_COUNT"
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
	return EncodeSpace(rangeKey)
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

func EncodeSpace(key string) string {
	if strings.Contains(key, "%20") {
		key = strings.ReplaceAll(key, "%20", DELIMITER)
	}

	if strings.Contains(key, " ") {
		key = strings.ReplaceAll(key, " ", DELIMITER)
	}
	return key
}

func DecodeSpace(key string) string {
	if strings.Contains(key, DELIMITER) {
		key = strings.ReplaceAll(key, DELIMITER, " ")
	}
	return key
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

func IsServiceEligibleForTag(streamData models.StreamData, rule models.RuleResponse) bool {

	fmt.Println("IsTagAttachable start")

	// ruleMetadataFieldList contains MetadataField and CoRuleMetadataField
	ruleMetadataFieldList := make([]string, 0)
	// ruleKeywordList contains Keyword and CoRuleKeyword
	ruleKeywordList := make([]string, 0)

	ruleMetadataFieldList = append(ruleMetadataFieldList, rule.MetadataField)
	ruleKeywordList = append(ruleKeywordList, rule.Keyword)

	if rule.CoRuleMetadataField != "" && rule.KeywordOperator != "" {
		ruleMetadataFieldList = append(ruleMetadataFieldList, rule.CoRuleMetadataField)
		ruleKeywordList = append(ruleKeywordList, rule.CoRuleKeyword)
	}

	if len(ruleMetadataFieldList) == 1 {
		fmt.Println("single condition")
	} else {
		fmt.Println("composite condition")
	}

	var cond [2]bool
	for i, ruleMetadataField := range ruleMetadataFieldList {
		cond[i] = matchCondition(ruleMetadataField, streamData, ruleKeywordList[i], rule.Operand, rule.RelationalOperator)
		fmt.Printf("Is :%v condition matched :%v\n", i, cond[i])
	}

	switch rule.KeywordOperator {
	case AND:
		fmt.Println("composite condition with AND: is true :", cond[0] && cond[1])
		return cond[0] && cond[1]
	case OR:
		fmt.Println("composite condition with OR: is true :", cond[0] || cond[1])
		return cond[0] || cond[1]
	case "":
		fmt.Println("single condition : is true :", cond[0])
		return cond[0]
	default:
		fmt.Println("unknown keyword operator")
		return false
	}
}

func matchCondition(ruleMetadataField string, streamData models.StreamData, keyword string, operand int, relationalOperator string) bool {
	// mdValue : description value, like value
	mdValue := getMetaDataFieldValue(ruleMetadataField, streamData)
	fmt.Println("matchCondition : metadata field value : ", mdValue)
	fmt.Printf("matchCondition : ruleMetadataField : %v : relationalOperator : %v: Is like == like %v\n", ruleMetadataField, relationalOperator, strings.ToLower(ruleMetadataField) == LIKE)

	if strings.ToLower(ruleMetadataField) == LIKE {
		likeCount, _ := strconv.Atoi(mdValue)
		switch relationalOperator {
		case GREATER_THAN:
			fmt.Printf("GREATER_THAN :Checking condition service_likeCount :%v rule like_count %v : matched : %v\n", likeCount, operand, likeCount > operand)
			if likeCount > operand {
				return true
			}
		case LESSER_THAN:
			fmt.Printf("LESSER_THAN :Checking condition service_likeCount :%v rule like_count %v : matched : %v\n", likeCount, operand, likeCount < operand)
			if likeCount < operand {
				return true
			}
		case GREATER_THAN_EQUAL:
			fmt.Printf("GREATER_THAN_EQUAL :Checking condition service_likeCount :%v rule like_count %v : matched : %v\n", likeCount, operand, likeCount >= operand)
			if likeCount >= operand {
				return true
			}
		case LESSER_THAN_EQUAL:
			fmt.Printf("LESSER_THAN_EQUAL :Checking condition service_likeCount :%v rule like_count %v : matched : %v\n", likeCount, operand, likeCount <= operand)
			if likeCount <= operand {
				return true
			}
		case EQUAL:
			fmt.Printf("EQUAL :Checking condition service_likeCount :%v rule like_count %v : matched : %v\n", likeCount, operand, likeCount == operand)
			if likeCount == operand {
				return true
			}
		}
	} else {
		keyword := strings.ToLower(keyword)
		fmt.Printf("CONTAIN : keyword %v : mdValue : %v", keyword, mdValue)
		if strings.Contains(mdValue, keyword) {
			fmt.Println("CONTAIN : found keyword in metadata field value")
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
	case PRICING:
		value = streamData.Pricing
	case BUSINESSMODEL:
		value = streamData.BusinessModel
	case DEPLOYMENT:
		value = streamData.Deployment
	case STAGE:
		value = streamData.Stage
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

func StreamDataToRuleConversion(streamData models.StreamData) (rule models.RuleResponse) {
	rule.PK = streamData.PK
	rule.SK = streamData.SK
	rule.RuleUUID = streamData.UUID
	rule.Operation = streamData.Operation
	rule.TagKey = streamData.TagKey
	rule.TagValue = streamData.TagValue
	rule.MetadataField = streamData.MetadataField
	rule.Keyword = streamData.Keyword
	rule.KeywordOperator = streamData.KeywordOperator
	rule.RelationalOperator = streamData.RelationalOperator
	rule.Operand = streamData.Operand
	rule.SubscriptionCount = streamData.SubscriptionCount
	rule.CoRuleMetadataField = streamData.CoRuleMetadataField
	rule.CoRuleKeyword = streamData.CoRuleKeyword

	return rule
}

func ServiceToStreamDataConversion(service models.ServiceResponse) (streamData models.StreamData) {
	streamData.PK = service.PK
	streamData.SK = service.SK
	streamData.UUID = service.ServiceUUID
	streamData.ServiceName = service.ServiceName
	streamData.Description = service.Description
	streamData.MoreAbout = service.MoreAbout
	streamData.Category = service.Category
	streamData.Like = service.Like
	streamData.Stage = service.Stage
	streamData.TargetSegment = service.TargetSegment
	streamData.Deployment = service.Deployment
	streamData.BusinessModel = service.BusinessModel
	streamData.Pricing = service.Pricing
	streamData.Location = service.Location

	return streamData
}
