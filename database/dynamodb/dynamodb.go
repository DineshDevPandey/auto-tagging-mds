package dynamodb

import (
	"errors"
	"fmt"
	"os"

	"github.com/auto-tagging-mds/database/models"
	"github.com/auto-tagging-mds/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type Database struct {
	db        *dynamodb.DynamoDB
	tableName models.Tables
}

var blank string = ""

func New(tablesName models.Tables) (*Database, error) {
	// Environment variables
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	region := os.Getenv("AWS_REGION")

	// DynamoDB
	sess := session.Must(session.NewSession())
	config := aws.NewConfig().WithRegion(region)
	if len(endpoint) > 0 {
		config = config.WithEndpoint(endpoint)
	}

	var db Database
	db.db = dynamodb.New(sess, config)
	db.tableName = tablesName

	return &db, nil
}

func (d *Database) CreateService(service models.ServiceRequest) (models.ServiceRequest, error) {

	// if its a fresh entry
	if service.ServiceUUID == "" {
		// check if the service already exists
		existService, err := d.GetService(service.ServiceName)
		if err != nil {
			return service, err
		}

		if existService.ServiceName != "" {
			return service, errors.New("Service already exist")
		}

		service.ServiceUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		service.CreatedAt, service.UpdatedAt = datetime, datetime
		service.PK = utils.GetPartitionKey(utils.SERVICE)
		service.SK = utils.GetRangeKey(utils.SERVICE, service.ServiceName, blank, blank, blank)
	}

	av, err := dynamodbattribute.MarshalMap(service)
	if err != nil {
		return service, err
	}
	if len(service.Category) == 0 {
		av = utils.NilToEmptySlice(av, "category")
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName.MDSTable),
	}

	_, err = d.db.PutItem(input)
	if err != nil {
		return service, err
	}

	return service, nil
}

func (d *Database) GetAllServices() ([]models.ServiceResponse, error) {

	services := []models.ServiceResponse{}
	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.SERVICE)

	keyCond := expression.Key(pkName).Equal(expression.Value(pkPrefix))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return services, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := d.db.Query(input)
	if err != nil {
		return services, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &services)
	if err != nil {
		return services, err
	}

	return services, nil
}

func (d *Database) GetService(name string) (models.ServiceResponse, error) {

	service := models.ServiceResponse{}
	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.SERVICE)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.SERVICE, name, blank, blank, blank)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	result, err := d.db.GetItem(input)
	if err != nil {
		return service, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &service)
	if err != nil {
		return service, err
	}

	return service, nil
}

func (d *Database) GetServiceByUUID(uuid string, projection *string) (models.ServiceResponse, error) {

	services := []models.ServiceResponse{}
	// pkName := utils.GetPartitionKeyName(utils.SERVICE)
	// pk := utils.GetPartitionKey(utils.SERVICE, name)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName.MDSTable),
		IndexName:              aws.String("uuid-index"),
		KeyConditionExpression: aws.String("#key = :value"),
		ProjectionExpression:   projection,
		ExpressionAttributeNames: map[string]*string{
			"#key": aws.String("uuid"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":value": {
				S: aws.String(uuid),
			},
		},
	}

	result, err := d.db.Query(input)
	if err != nil {
		return models.ServiceResponse{}, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &services)
	if err != nil {
		return models.ServiceResponse{}, err
	}

	if len(result.Items) > 0 {
		return services[0], nil
	}

	return models.ServiceResponse{}, nil
}

func (d *Database) UpdateService(updatedService models.ServiceRequest, serviceUUID string) error {

	oldService, err := d.GetServiceByUUID(serviceUUID, nil)
	if err != nil {
		return err
	}

	if oldService.ServiceName == "" {
		return errors.New("service not found")
	}

	oldServiceName := utils.GetRangeKey(utils.SERVICE, oldService.ServiceName, blank, blank, blank)
	newServiceName := utils.GetRangeKey(utils.SERVICE, updatedService.ServiceName, blank, blank, blank)

	// if service name is changed, delete old entry and create new one
	if oldServiceName != newServiceName {
		err := d.DeleteService(oldService.ServiceName)
		if err != nil {
			return err
		}
	}
	// old created at
	updatedService.CreatedAt = oldService.CreatedAt
	// new updated at
	updatedService.UpdatedAt = utils.DateString("datetime")
	updatedService.PK = utils.GetPartitionKey(utils.SERVICE)
	updatedService.SK = utils.GetRangeKey(utils.SERVICE, updatedService.ServiceName, blank, blank, blank)
	_, err = d.CreateService(updatedService)
	if err != nil {
		// TODO: restore old entry in case of error
		return err
	}

	return nil
}

func (d *Database) DeleteService(name string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.SERVICE)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.SERVICE, name, blank, blank, blank)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	// GetItem from dynamodb table
	_, err := d.db.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

// ################################################################################################# Company

func (d *Database) CreateCompany(company models.CompanyRequest) error {

	// if its a fresh entry
	if company.CompanyUUID == "" {
		// check if companyalready exist
		existCompany, err := d.GetCompany(company.CompanyName)
		if err != nil {
			return err
		}

		if existCompany.CompanyName != "" {
			return errors.New("Company already exist")
		}

		company.CompanyUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		company.CreatedAt, company.UpdatedAt = datetime, datetime
		company.PK = utils.GetPartitionKey(utils.COMPANY)
		company.SK = utils.GetRangeKey(utils.COMPANY, company.CompanyName, blank, blank, blank)
	}

	av, err := dynamodbattribute.MarshalMap(company)
	if err != nil {
		return err
	}

	if len(company.ServiceList) == 0 {
		av = utils.NilToEmptySlice(av, "service_list")
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName.MDSTable),
	}

	_, err = d.db.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetAllCompanies() ([]models.CompanyResponse, error) {

	companies := []models.CompanyResponse{}
	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.COMPANY)

	keyCond := expression.Key(pkName).Equal(expression.Value(pkPrefix))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return companies, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := d.db.Query(input)
	if err != nil {
		return companies, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &companies)
	if err != nil {
		return companies, err
	}

	projection := aws.String("service_name")
	// get latest service name
	for cindex, company := range companies {
		for sindex, srv := range company.ServiceList {
			if srv.ServiceUUID != "" {
				service, err := d.GetServiceByUUID(srv.ServiceUUID, projection)
				if err == nil {
					companies[cindex].ServiceList[sindex].ServiceUUID = srv.ServiceUUID
					companies[cindex].ServiceList[sindex].ServiceName = service.ServiceName
				}
			}
		}
	}

	return companies, nil
}

func (d *Database) GetCompany(name string) (models.CompanyResponse, error) {

	company := models.CompanyResponse{}
	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.COMPANY)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.COMPANY, name, blank, blank, blank)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	result, err := d.db.GetItem(input)
	if err != nil {
		return company, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &company)
	if err != nil {
		return company, err
	}

	projection := aws.String("service_name")
	// get latest service name
	for sindex, srv := range company.ServiceList {
		if srv.ServiceUUID != "" {
			service, err := d.GetServiceByUUID(srv.ServiceUUID, projection)
			if err == nil {
				company.ServiceList[sindex].ServiceUUID = srv.ServiceUUID
				company.ServiceList[sindex].ServiceName = service.ServiceName
			}
		}
	}

	return company, nil
}

func (d *Database) GetCompanyByUUID(uuid string) (models.CompanyResponse, error) {

	companies := []models.CompanyResponse{}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName.MDSTable),
		IndexName:              aws.String("uuid-index"),
		KeyConditionExpression: aws.String("#key = :value"),
		ExpressionAttributeNames: map[string]*string{
			"#key": aws.String("uuid"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":value": {
				S: aws.String(uuid),
			},
		},
	}

	result, err := d.db.Query(input)
	if err != nil {
		return models.CompanyResponse{}, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &companies)
	if err != nil {
		return models.CompanyResponse{}, err
	}

	if len(result.Items) > 0 {
		return companies[0], nil
	}

	return models.CompanyResponse{}, nil
}

func (d *Database) UpdateCompany(updatedCompany models.CompanyRequest, companyUUID string) error {

	oldCompany, err := d.GetCompanyByUUID(companyUUID)
	if err != nil {
		return err
	}

	if oldCompany.CompanyName == "" {
		return errors.New("company not found")
	}

	oldCompanyName := utils.GetRangeKey(utils.COMPANY, oldCompany.CompanyName, blank, blank, blank)
	newCompanyName := utils.GetRangeKey(utils.COMPANY, updatedCompany.CompanyName, blank, blank, blank)

	// if company name is changed, delete old entry and create new one
	if oldCompanyName != newCompanyName {
		err := d.DeleteCompany(oldCompany.CompanyName)
		if err != nil {
			return err
		}
	}
	// old created at
	updatedCompany.CreatedAt = oldCompany.CreatedAt
	// new updated at
	updatedCompany.UpdatedAt = utils.DateString("datetime")
	updatedCompany.PK = utils.GetPartitionKey(utils.COMPANY)
	updatedCompany.SK = utils.GetRangeKey(utils.COMPANY, updatedCompany.CompanyName, blank, blank, blank)
	err = d.CreateCompany(updatedCompany)
	if err != nil {
		// TODO: restore old entry in case of error
		return err
	}

	return nil
}

func (d *Database) DeleteCompany(name string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.COMPANY)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.COMPANY, name, blank, blank, blank)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	// GetItem from dynamodb table
	_, err := d.db.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

// ################################################################################################# Tag

// In DB
// TG#Keyword#value1
// TG#Keyword#value2
// TG#Keyword#value3
func (d *Database) CreateTag(tag models.TagCreateRequest) (models.TagCreateRequest, error) {

	// check if the service already exists
	existTag, err := d.GetTag(tag.Key, tag.Value)
	if err != nil {
		return tag, err
	}

	if existTag.Key != "" {
		return tag, errors.New("Tag already exist")
	}

	datetime := utils.DateString("datetime")
	tag.CreatedAt, tag.UpdatedAt = datetime, datetime
	tag.PK = utils.GetPartitionKey(utils.TAG)
	tag.SK = utils.GetRangeKey(utils.TAG, tag.Key, tag.Value, blank, blank)

	av, err := dynamodbattribute.MarshalMap(tag)
	if err != nil {
		return tag, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName.MDSTable),
	}

	_, err = d.db.PutItem(input)
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func (d *Database) GetAllTags() ([]models.TagListResponse, error) {

	tags := []models.TagResponse{}
	tagList := make([]models.TagListResponse, 0)

	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.TAG)

	keyCond := expression.Key(pkName).Equal(expression.Value(pkPrefix))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return tagList, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := d.db.Query(input)
	if err != nil {
		return tagList, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &tags)
	if err != nil {
		return tagList, err
	}

	return createTagResponse(tags, tagList), nil
}

func createTagResponse(tags []models.TagResponse, tagList []models.TagListResponse) []models.TagListResponse {
	tagMap := make(map[string][]string, 0)
	createdAt := ""
	updatedAt := ""

	for _, tag := range tags {
		tagMap[tag.Key] = append(tagMap[tag.Key], tag.Value)
		// TODO: create logic to get oldest created_at and latest updated_at
		createdAt = tag.CreatedAt
		updatedAt = tag.UpdatedAt
	}

	for key, value := range tagMap {
		temp := models.TagListResponse{
			Key:       key,
			Values:    value,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		tagList = append(tagList, temp)
	}

	return tagList
}

func (d *Database) DeleteTag(key string, value string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.TAG)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.TAG, key, value, blank, blank)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	_, err := d.db.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetTag(key string, value string) (models.TagListResponse, error) {

	tags := []models.TagResponse{}
	resultTag := []models.TagListResponse{}
	dummy := models.TagListResponse{}

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.TAG)

	skName := utils.GetRangeKeyName()
	var sk string
	if value == "" {
		sk = utils.GetRangeKey(utils.TAG, key, blank, blank, blank)
	} else {
		sk = utils.GetRangeKey(utils.TAG, key, value, blank, blank)
	}

	keyCond := expression.KeyAnd(expression.Key(pkName).Equal(expression.Value(pk)), expression.Key(skName).BeginsWith(sk))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return dummy, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := d.db.Query(input)
	if err != nil {
		return dummy, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &tags)
	if err != nil {
		return dummy, err
	}

	resp := createTagResponse(tags, resultTag)
	if len(resp) > 0 {
		return createTagResponse(tags, resultTag)[0], nil
	}

	return dummy, nil
}

// ################################################################################################# Rule

func (d *Database) CreateRule(rule models.RuleRequest) error {

	if rule.RuleUUID == "" {
		// check if rule already exist
		existRule, err := d.GetRule(rule.RuleUUID)
		if err != nil {
			return err
		}

		if existRule.TagKey != "" {
			return errors.New("Rule already exist")
		}

		rule.RuleUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		rule.CreatedAt, rule.UpdatedAt = datetime, datetime
		rule.PK = utils.GetPartitionKey(utils.RULE)
		rule.SK = utils.GetRangeKey(utils.RULE, rule.TagKey, rule.TagValue, rule.MetadataField, rule.Operation)
	}

	// check if companyalready exist
	existRule, err := d.GetRule(rule.RuleUUID)
	if err != nil {
		return err
	}

	if existRule.RuleUUID != "" {
		return errors.New("Rule already exist")
	}

	rule.RuleUUID = utils.GetUUID()
	datetime := utils.DateString("datetime")
	rule.CreatedAt, rule.UpdatedAt = datetime, datetime
	rule.PK = utils.GetPartitionKey(utils.RULE)
	rule.SK = utils.GetRangeKey(utils.RULE, rule.TagKey, rule.TagValue, rule.MetadataField, rule.Operation)

	av, err := dynamodbattribute.MarshalMap(rule)
	if err != nil {
		return err
	}

	if len(rule.Keyword) == 0 {
		av = utils.NilToEmptySlice(av, "keyword")
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName.MDSTable),
	}

	_, err = d.db.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetAllRules() ([]models.RuleResponse, error) {

	rules := []models.RuleResponse{}
	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.RULE)

	keyCond := expression.Key(pkName).Equal(expression.Value(pkPrefix))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return rules, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := d.db.Query(input)
	if err != nil {
		return rules, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &rules)
	if err != nil {
		return rules, err
	}

	return rules, nil
}

// complete it
func (d *Database) GetRule(ruleUUID string) (models.RuleResponse, error) {

	rule, err := d.GetRuleByUUID(ruleUUID)
	if err != nil {
		return rule, err
	}

	return rule, nil
}

// complete it
func (d *Database) UpdateRule(updatedRule models.RuleRequest, ruleUUID string) error {

	oldRule, err := d.GetRule(ruleUUID)
	if err != nil {
		return err
	}

	if oldRule.Operation == "" {
		return errors.New("rule not found")
	}

	// old created at
	updatedRule.CreatedAt = oldRule.CreatedAt
	// new updated at
	updatedRule.UpdatedAt = utils.DateString("datetime")
	updatedRule.PK = utils.GetPartitionKey(utils.RULE)
	updatedRule.SK = utils.GetRangeKey(utils.RULE, updatedRule.RuleUUID, blank, blank, blank)
	err = d.CreateRule(updatedRule)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) DeleteRule(ruleUUID string) error {

	rule, err := d.GetRuleByUUID(ruleUUID)
	if err != nil {
		return err
	}

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.RULE)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.RULE, rule.TagKey, rule.TagValue, rule.MetadataField, rule.Operation)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
			skName: {
				S: aws.String(sk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	// GetItem from dynamodb table
	_, err = d.db.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetRuleByUUID(uuid string) (models.RuleResponse, error) {

	rules := []models.RuleResponse{}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName.MDSTable),
		IndexName:              aws.String("uuid-index"),
		KeyConditionExpression: aws.String("#key = :value"),
		ExpressionAttributeNames: map[string]*string{
			"#key": aws.String("uuid"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":value": {
				S: aws.String(uuid),
			},
		},
	}

	result, err := d.db.Query(input)
	if err != nil {
		return models.RuleResponse{}, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &rules)
	if err != nil {
		return models.RuleResponse{}, err
	}

	if len(result.Items) > 0 {
		return rules[0], nil
	}

	return models.RuleResponse{}, nil
}

func (d *Database) AttachTagWithService(streamData models.StreamData, rules []models.RuleResponse) error {

	for i, rule := range rules {
		fmt.Printf("rule number : %v : key : %v : value : %v\n", i, rule.TagKey, rule.TagValue)

		updateDb := false
		switch rule.Operation {
		case "CONTAIN":
			fallthrough
		case "RELATION":
			updateDb = utils.IsTagValueFound(streamData, rule)
		case "SUBSCRIPTION_COUNT":
			// TODO: create logic for condition

		}
		_ = updateDb
		// if updateDb {
		// cat := models.Category{Key: rule.TagKey, Value: rule.TagValue}
		// 	service.Category = utils.AppendTag(streamData.Category, cat)
		// 	fmt.Printf("service.Category : %v\n", streamData.Category)
		// 	d.CreateService(streamData)
		// 	fmt.Println("streamData updated")
		// }
	}
	return nil
}

func (d *Database) UpdateCategoryInService(cat models.Category, streamData models.StreamData) error {

	// construct the UpdateItemInput struct
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName.MDSTable),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(streamData.PK),
			},
			"SK": {
				S: aws.String(streamData.SK),
			},
		},
		UpdateExpression: aws.String("SET #attr = list_append(#attr, :val)"),
		ExpressionAttributeNames: map[string]*string{
			"#attr": aws.String("category"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": {
				L: []*dynamodb.AttributeValue{
					{
						M: map[string]*dynamodb.AttributeValue{
							"key":   {S: aws.String(cat.Key)},
							"value": {S: aws.String(cat.Value)},
						},
					},
				},
			},
		},
	}

	// call the UpdateItem method to update the item in the table
	_, err := d.db.UpdateItem(updateInput)

	return err
}
