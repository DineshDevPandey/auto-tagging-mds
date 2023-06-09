package models

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Tables struct {
	MDSTable string
}

type DynamoDBEvent struct {
	Records []DynamoDBEventRecord `json:"Records"`
}

type DynamoDBEventRecord struct {
	AWSRegion      string                       `json:"awsRegion"`
	Change         DynamoDBStreamRecord         `json:"dynamodb"`
	EventID        string                       `json:"eventID"`
	EventName      string                       `json:"eventName"`
	EventSource    string                       `json:"eventSource"`
	EventVersion   string                       `json:"eventVersion"`
	EventSourceArn string                       `json:"eventSourceARN"`
	UserIdentity   *events.DynamoDBUserIdentity `json:"userIdentity,omitempty"`
}

type DynamoDBStreamRecord struct {
	ApproximateCreationDateTime events.SecondsEpochTime `json:"ApproximateCreationDateTime,omitempty"`
	// changed to map[string]*dynamodb.AttributeValue
	Keys map[string]*dynamodb.AttributeValue `json:"Keys,omitempty"`
	// changed to map[string]*dynamodb.AttributeValue
	NewImage map[string]*dynamodb.AttributeValue `json:"NewImage,omitempty"`
	// changed to map[string]*dynamodb.AttributeValue
	OldImage       map[string]*dynamodb.AttributeValue `json:"OldImage,omitempty"`
	SequenceNumber string                              `json:"SequenceNumber"`
	SizeBytes      int64                               `json:"SizeBytes"`
	StreamViewType string                              `json:"StreamViewType"`
}

type Service struct {
	ServiceName   string     `json:"service_name"`
	Description   string     `json:"description"`
	MoreAbout     string     `json:"more_about"`
	Category      []Category `json:"category"`
	Like          int        `json:"like"`
	Stage         string     `json:"stage"`
	TargetSegment string     `json:"target_segment"`
	Deployment    string     `json:"deployment"`
	BusinessModel string     `json:"business_model"`
	Pricing       string     `json:"pricing"`
	Location      string     `json:"location"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

type Category struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ServiceUUID auto generated and used to update service data.
// If service name is updated by user, we will not be able to update it in DB.
// So we will delete onld entry and insert updated entry with same uuid
type ServiceRequest struct {
	PK            string     `json:"PK"`   //auto generated by BE
	SK            string     `json:"SK"`   //auto generated by BE
	ServiceUUID   string     `json:"uuid"` //auto generated by BE for create, sent by FE for update
	ServiceName   string     `json:"service_name" validate:"min=1,required"`
	Description   string     `json:"description"`
	MoreAbout     string     `json:"more_about"`
	Category      []Category `json:"category"`
	Like          int        `json:"like"`
	Stage         string     `json:"stage"`
	TargetSegment string     `json:"target_segment"`
	Deployment    string     `json:"deployment"`
	BusinessModel string     `json:"business_model"`
	Pricing       string     `json:"pricing"`
	Location      string     `json:"location"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

type ServiceResponse struct {
	PK            string     `json:"PK"` //auto generated by BE
	SK            string     `json:"SK"` //auto generated by BE
	ServiceUUID   string     `json:"uuid"`
	ServiceName   string     `json:"service_name"`
	Description   string     `json:"description"`
	MoreAbout     string     `json:"more_about"`
	Category      []Category `json:"category"`
	Like          int        `json:"like"`
	Stage         string     `json:"stage"`
	TargetSegment string     `json:"target_segment"`
	Deployment    string     `json:"deployment"`
	BusinessModel string     `json:"business_model"`
	Pricing       string     `json:"pricing"`
	Location      string     `json:"location"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

type Company struct {
	PK          string   `json:"PK"` //auto generated
	SK          string   `json:"SK"` //auto generated by BE
	CompanyUUID string   `json:"uuid"`
	CompanyName string   `json:"company_name" validate:"min=1,required"`
	Description string   `json:"description"`
	ServiceList []string `json:"service_list"` // holds service uuid
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CompanyRequest struct {
	PK          string   `json:"PK"`   //auto generated
	SK          string   `json:"SK"`   //auto generated by BE
	CompanyUUID string   `json:"uuid"` //auto generated by BE for create, sent by FE for update
	CompanyName string   `json:"company_name" validate:"min=1,required"`
	Description string   `json:"description"`
	ServiceList []string `json:"service_list"` // FE sends only service uuid list
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CompanyResponse struct {
	CompanyUUID string     `json:"uuid"`
	CompanyName string     `json:"company_name"`
	Description string     `json:"description"`
	ServiceList []Services `json:"service_list"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type Services struct {
	ServiceUUID string `json:"uuid"`
	ServiceName string `json:"service_name"`
}

type TagRequest struct {
	PK        string `json:"PK"` //auto generated
	SK        string `json:"SK"` //auto generated by BE
	Key       string `json:"key" validate:"min=1,required"`
	Value     string `json:"value" validate:"min=1,required"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TagCreateRequest struct {
	PK        string `json:"PK"`                              //auto generated
	SK        string `json:"SK"`                              //auto generated by BE
	Key       string `json:"key" validate:"min=1,required"`   // need(1/2)
	Value     string `json:"value" validate:"min=1,required"` // need(2/2), add 1 at a time
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TagResponse struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TagListResponse struct {
	Key       string   `json:"key"`
	Values    []string `json:"values"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type Tags struct {
	TagValueUUID string `json:"uuid"`
	TagValue     string `json:"tag_name"`
}

type Rule struct {
	Operation           string `json:"operation" validate:"min=1,required"` // CONTAIN/RELATION
	TagKey              string `json:"tag_key"`
	TagValue            string `json:"tag_value"`
	MetadataField       string `json:"metadata_field"`
	Keyword             string `json:"keyword"`
	KeywordOperator     string `json:"keyword_operator"`
	Operator            string `json:"relational_operator"`
	Operand             int    `json:"relational_operand"`
	SubscriptionCount   int    `json:"subscription_count"`
	IsSibling           bool   `json:"is_sibling"`
	SiblingUUID         string `json:"sibling_uuid"`
	CoRuleMetadataField string `json:"corule_metadata_field"`
	CoRuleKeyword       string `json:"corule_keyword"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type RuleRequest struct {
	PK                  string `json:"PK"` //auto generated
	SK                  string `json:"SK"` //auto generated by BE
	RuleUUID            string `json:"uuid"`
	Operation           string `json:"operation" validate:"required"` //CONTAIN|RELATION|SUBSCRIPTION_COUNT
	TagKey              string `json:"tag_key" validate:"min=1"`
	TagValue            string `json:"tag_value" validate:"min=1"`
	MetadataField       string `json:"metadata_field"`
	Keyword             string `json:"keyword"`
	KeywordOperator     string `json:"keyword_operator"`    //AND|OR
	RelationalOperator  string `json:"relational_operator"` //GREATER_THAN|LESSER_THAN|EQUAL|GREATER_THAN_EQUAL|LESSER_THAN_EQUAL
	Operand             int    `json:"relational_operand"`
	SubscriptionCount   int    `json:"subscription_count"`
	CoRuleMetadataField string `json:"corule_metadata_field"`
	CoRuleKeyword       string `json:"corule_keyword"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type RuleResponse struct {
	PK                  string `json:"PK"` //auto generated
	SK                  string `json:"SK"` //auto generated by BE
	RuleUUID            string `json:"uuid"`
	Operation           string `json:"operation"`
	TagKey              string `json:"tag_key"`
	TagValue            string `json:"tag_value"`
	MetadataField       string `json:"metadata_field"`
	Keyword             string `json:"keyword"`
	KeywordOperator     string `json:"keyword_operator"`
	RelationalOperator  string `json:"relational_operator"`
	Operand             int    `json:"relational_operand"`
	SubscriptionCount   int    `json:"subscription_count"`
	CoRuleMetadataField string `json:"corule_metadata_field"`
	CoRuleKeyword       string `json:"corule_keyword"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type StreamData struct {
	PK                  string     `json:"PK"`
	SK                  string     `json:"SK"`
	UUID                string     `json:"uuid,omitempty"`
	Operation           string     `json:"operation,omitempty"`
	TagKey              string     `json:"tag_key,omitempty"`
	TagValue            string     `json:"tag_value,omitempty"`
	MetadataField       string     `json:"metadata_field,omitempty"`
	Keyword             string     `json:"keyword,omitempty"`
	KeywordOperator     string     `json:"keyword_operator,omitempty"`
	RelationalOperator  string     `json:"relational_operator,omitempty"`
	Operand             int        `json:"relational_operand,omitempty"`
	SubscriptionCount   int        `json:"subscription_count,omitempty"`
	CoRuleMetadataField string     `json:"corule_metadata_field"`
	CoRuleKeyword       string     `json:"corule_keyword"`
	Key                 string     `json:"key,omitempty"`
	Value               string     `json:"value,omitempty"`
	CompanyName         string     `json:"company_name,omitempty"`
	Description         string     `json:"description,omitempty"`
	ServiceList         []string   `json:"service_list,omitempty"`
	ServiceName         string     `json:"service_name,omitempty"`
	MoreAbout           string     `json:"more_about,omitempty"`
	Category            []Category `json:"category,omitempty"`
	Like                int        `json:"like,omitempty"`
	Stage               string     `json:"stage,omitempty"`
	TargetSegment       string     `json:"target_segment,omitempty"`
	Deployment          string     `json:"deployment,omitempty"`
	BusinessModel       string     `json:"business_model,omitempty"`
	Pricing             string     `json:"pricing,omitempty"`
	Location            string     `json:"location,omitempty"`
}
