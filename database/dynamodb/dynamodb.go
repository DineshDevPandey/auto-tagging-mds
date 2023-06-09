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

func (d *Database) IsTagValid(key, value string) (bool, error) {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.TAG)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.TAG, key, value, blank)

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
		return false, err
	}

	tag := models.TagResponse{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &tag)
	if err != nil {
		return false, err
	}

	if tag.Key != "" {
		return true, nil
	}

	return false, nil
}

func (d *Database) VerifyTag(category []models.Category) error {
	for _, cat := range category {
		valid, err := d.IsTagValid(cat.Key, cat.Value)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New(fmt.Sprintf("Invalid tag : (%v:%v)", cat.Key, cat.Value))
		}
	}
	return nil
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
		service.SK = utils.GetRangeKey(utils.SERVICE, service.ServiceName, blank, blank)
	}

	err := d.VerifyTag(service.Category)
	if err != nil {
		return service, err
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
	sk := utils.GetRangeKey(utils.SERVICE, name, blank, blank)

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

	oldServiceName := utils.GetRangeKey(utils.SERVICE, oldService.ServiceName, blank, blank)
	newServiceName := utils.GetRangeKey(utils.SERVICE, updatedService.ServiceName, blank, blank)

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
	updatedService.SK = utils.GetRangeKey(utils.SERVICE, updatedService.ServiceName, blank, blank)

	err = d.VerifyTag(updatedService.Category)
	if err != nil {
		return err
	}

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
	sk := utils.GetRangeKey(utils.SERVICE, name, blank, blank)

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

func (d *Database) VerifyService(serviceList []string) (bool, error) {
	projection := aws.String("service_name")
	for _, serviceId := range serviceList {
		s, err := d.GetServiceByUUID(serviceId, projection)
		if err != nil {
			return false, err
		}
		if s.ServiceName == "" {
			return false, errors.New(fmt.Sprintf("Service %s not found", serviceId))
		}
	}
	return true, nil
}

func (d *Database) CreateCompany(company models.CompanyRequest) (models.CompanyRequest, error) {

	// if its a fresh entry
	if company.CompanyUUID == "" {
		// check if companyalready exist
		existCompany, err := d.GetCompany(company.CompanyName)
		if err != nil {
			return company, err
		}

		if existCompany.CompanyName != "" {
			return company, errors.New("Company already exist")
		}

		company.CompanyUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		company.CreatedAt, company.UpdatedAt = datetime, datetime
		company.PK = utils.GetPartitionKey(utils.COMPANY)
		company.SK = utils.GetRangeKey(utils.COMPANY, company.CompanyName, blank, blank)
	}

	valid, err := d.VerifyService(company.ServiceList)
	if err != nil {
		return company, err
	}

	if !valid {
		return company, err
	}

	av, err := dynamodbattribute.MarshalMap(company)
	if err != nil {
		return company, err
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
		return company, err
	}

	return company, nil
}

func (d *Database) GetAllCompanies() ([]models.CompanyResponse, error) {

	companies := make([]models.CompanyResponse, 0)
	companiesTemp := []models.Company{}

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

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &companiesTemp)
	if err != nil {
		return companies, err
	}

	projection := aws.String("service_name")
	// get latest service name
	for _, company := range companiesTemp {

		temp := models.CompanyResponse{}
		s := make([]models.Services, 0)
		for _, srvId := range company.ServiceList {
			service, err := d.GetServiceByUUID(srvId, projection)
			if err == nil {
				s = append(s, models.Services{ServiceUUID: srvId, ServiceName: service.ServiceName})
			}
		}
		temp.CompanyUUID = company.CompanyUUID
		temp.CompanyName = company.CompanyName
		temp.CreatedAt = company.CreatedAt
		temp.UpdatedAt = company.UpdatedAt
		temp.Description = company.Description
		temp.ServiceList = s
		companies = append(companies, temp)
	}

	return companies, nil
}

func (d *Database) GetCompany(name string) (models.CompanyResponse, error) {

	company := models.CompanyResponse{}
	companyTemp := models.Company{}

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.COMPANY)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.COMPANY, name, blank, blank)

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

	err = dynamodbattribute.UnmarshalMap(result.Item, &companyTemp)
	if err != nil {
		return company, err
	}

	projection := aws.String("service_name")
	// get latest service name
	s := make([]models.Services, 0)
	for _, srvId := range companyTemp.ServiceList {
		service, err := d.GetServiceByUUID(srvId, projection)
		if err == nil {
			s = append(s, models.Services{ServiceUUID: srvId, ServiceName: service.ServiceName})
		}
	}
	company.CompanyUUID = companyTemp.CompanyUUID
	company.CompanyName = companyTemp.CompanyName
	company.CreatedAt = companyTemp.CreatedAt
	company.UpdatedAt = companyTemp.UpdatedAt
	company.Description = companyTemp.Description
	company.ServiceList = s

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

	oldCompanyName := utils.GetRangeKey(utils.COMPANY, oldCompany.CompanyName, blank, blank)
	newCompanyName := utils.GetRangeKey(utils.COMPANY, updatedCompany.CompanyName, blank, blank)

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
	updatedCompany.SK = utils.GetRangeKey(utils.COMPANY, updatedCompany.CompanyName, blank, blank)

	valid, err := d.VerifyService(updatedCompany.ServiceList)
	if err != nil {
		return err
	}

	if !valid {
		return err
	}

	_, err = d.CreateCompany(updatedCompany)
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
	sk := utils.GetRangeKey(utils.COMPANY, name, blank, blank)

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
	tag.SK = utils.GetRangeKey(utils.TAG, tag.Key, tag.Value, blank)

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
	sk := utils.GetRangeKey(utils.TAG, key, value, blank)

	fmt.Println("DeleteTag sk :", sk)

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
		sk = utils.GetRangeKey(utils.TAG, key, blank, blank)
	} else {
		sk = utils.GetRangeKey(utils.TAG, key, value, blank)
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

func (d *Database) IsDuplicateRule(rule models.RuleRequest) (bool, error) {
	keyCond := expression.Key(utils.GetPartitionKeyName()).Equal(expression.Value(utils.GetPartitionKey(utils.RULE)))
	filter1 := expression.Name("operation").Equal(expression.Value(rule.Operation)).
		And(expression.Name("tag_key").Equal(expression.Value(rule.TagKey))).
		And(expression.Name("metadata_field").Equal(expression.Value(rule.MetadataField))).
		And(expression.Name("keyword").Equal(expression.Value(rule.Keyword))).
		And(expression.Name("keyword_operator").Equal(expression.Value(rule.KeywordOperator))).
		And(expression.Name("relational_operator").Equal(expression.Value(rule.RelationalOperator))).
		And(expression.Name("subscription_count").Equal(expression.Value(rule.SubscriptionCount))).
		And(expression.Name("relational_operand").Equal(expression.Value(rule.Operand))).
		And(expression.Name("corule_metadata_field").Equal(expression.Value(rule.CoRuleMetadataField))).
		And(expression.Name("corule_keyword").Equal(expression.Value(rule.CoRuleKeyword))).
		And(expression.Name("tag_value").Equal(expression.Value(rule.TagValue)))

	expr, err := expression.NewBuilder().WithFilter(filter1).WithKeyCondition(keyCond).Build()
	if err != nil {
		fmt.Printf("Expression builder error : %v\n", err)
		return false, err
	}

	// input for GetItem
	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      aws.String(utils.GetPartitionKeyName()),
	}

	// GetItem from dynamodb table
	result, err := d.db.Query(input)
	if err != nil {
		return false, err
	}

	item := []models.RuleRequest{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &item)
	if err != nil {
		return false, err
	}

	if len(item) > 0 {
		return true, nil
	}
	return false, nil
}

func (d *Database) CreateRule(rule models.RuleRequest) (models.RuleRequest, error) {
	// check if rule already exist
	isDuplicateRule, err := d.IsDuplicateRule(rule)
	if err != nil {
		return rule, err
	}

	if isDuplicateRule {
		return rule, errors.New("Rule already exist")
	}

	rule.RuleUUID = utils.GetUUID()
	datetime := utils.DateString("datetime")
	rule.CreatedAt, rule.UpdatedAt = datetime, datetime
	rule.PK = utils.GetPartitionKey(utils.RULE)
	rule.SK = utils.GetRangeKey(utils.RULE, blank, blank, rule.RuleUUID)

	cat := make([]models.Category, 0)
	cat = append(cat, models.Category{Key: rule.TagKey, Value: rule.TagValue})
	err = d.VerifyTag(cat)
	if err != nil {
		return rule, err
	}

	err = d.insertRule(rule)
	if err != nil {
		return rule, err
	}

	return rule, nil
}

func (d *Database) insertRule(rule models.RuleRequest) error {
	av, err := dynamodbattribute.MarshalMap(rule)
	if err != nil {
		return err
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

func (d *Database) GetRule(ruleUUID string) (models.RuleResponse, error) {

	rule := models.RuleResponse{}
	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.RULE)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.RULE, blank, blank, ruleUUID)

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
		return rule, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &rule)
	if err != nil {
		return rule, err
	}

	return rule, nil
}

// send both values togather
func (d *Database) UpdateRule(updatedRule models.RuleRequest, ruleUUID string) error {

	oldRule, err := d.GetRule(ruleUUID)
	if err != nil {
		return err
	}

	if oldRule.Operation == "" {
		return errors.New("rule not found")
	}

	// old created at
	updatedRule.PK = oldRule.PK
	updatedRule.SK = oldRule.SK
	updatedRule.CreatedAt = oldRule.CreatedAt
	// new updated at
	updatedRule.UpdatedAt = utils.DateString("datetime")

	cat := make([]models.Category, 0)
	cat = append(cat, models.Category{Key: updatedRule.TagKey, Value: updatedRule.TagValue})
	err = d.VerifyTag(cat)
	if err != nil {
		return err
	}

	err = d.insertRule(updatedRule)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) DeleteRule(ruleUUID string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.RULE)

	skName := utils.GetRangeKeyName()
	sk := utils.GetRangeKey(utils.RULE, blank, blank, ruleUUID)

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

func (d *Database) IsServiceEligibleForTag(streamData models.StreamData, rule models.RuleResponse) (bool, error) {

	keyCond := expression.Key(utils.GetPartitionKeyName()).Equal(expression.Value(utils.GetPartitionKey(utils.COMPANY)))
	filter := expression.Contains(expression.Name("service_list"), streamData.UUID)

	expr, err := expression.NewBuilder().WithFilter(filter).WithKeyCondition(keyCond).Build()
	if err != nil {
		fmt.Printf("Expression builder error : %v\n", err)
		return false, err
	}

	// input for GetItem
	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName.MDSTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      aws.String("company_name"),
	}

	// GetItem from dynamodb table
	result, err := d.db.Query(input)
	if err != nil {
		return false, err
	}

	item := []models.CompanyResponse{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &item)
	if err != nil {
		return false, err
	}

	fmt.Println("No of companies : ", len(item))
	fmt.Println("SubscriptionCount : ", rule.SubscriptionCount)
	if rule.SubscriptionCount < len(item) {
		return true, nil
	}
	return false, err
}

// execute when new service is created, here streamData contains service data
func (d *Database) AttachTagWithService(streamData models.StreamData, rules []models.RuleResponse) error {

	for i, rule := range rules {
		fmt.Printf("rule number : %v : key : %v : value : %v\n", i+1, rule.TagKey, rule.TagValue)
		updateDb := false
		var err error

		switch rule.Operation {
		case utils.CONTAIN:
			fallthrough

		case utils.RELATION:
			updateDb = utils.IsServiceEligibleForTag(streamData, rules[i])
			fmt.Println("IsServiceEligibleForTag :", updateDb)

		case utils.SUBSCRIPTION_COUNT:
			// check if this service is subscribe for more than subscription threshold
			updateDb, err = d.IsServiceEligibleForTag(streamData, rule)
			if err != nil {
				return err
			}
			fmt.Println("IsServiceEligibleForTag :", updateDb)
		}

		if updateDb {
			d.UpdateTagToService(streamData, rule)
		}
	}
	return nil
}

func (d *Database) AppendTagToService(cat models.Category, streamData models.StreamData) error {

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

// here streamData contains service data
func (d *Database) UpdateTagToService(streamData models.StreamData, rule models.RuleResponse) error {
	cat := models.Category{Key: rule.TagKey, Value: rule.TagValue}
	if isPresent := utils.IsTagAlreadyPresent(streamData.Category, cat); isPresent {
		fmt.Printf("tag already present : key : %v : value : %v\n", cat.Key, cat.Value)
		return nil
	}

	d.AppendTagToService(cat, streamData)
	fmt.Println("streamData updated : cat : ", cat)
	return nil
}

// execute when new rule is created, here streamData contains rule
func (d *Database) ProcessRuleForServices(streamData models.StreamData, services []models.ServiceResponse) error {
	fmt.Println("start ProcessRuleForServices")
	rule := utils.StreamDataToRuleConversion(streamData)

	rules := make([]models.RuleResponse, 0)
	rules = append(rules, rule)
	for _, service := range services {
		stData := utils.ServiceToStreamDataConversion(service)

		if rule.Operation == utils.SUBSCRIPTION_COUNT {
			// check if this service is subscribe for more than subscription threshold
			updateDb, err := d.IsServiceEligibleForTag(stData, rule)
			if err != nil {
				return err
			}
			if updateDb {
				d.UpdateTagToService(streamData, rule)
			}
		}

		err := d.AttachTagWithService(stData, rules)
		if err != nil {
			return err
		}
	}
	return nil
}

// here stream data contains company data
func (d *Database) UpdateServiceTagForSubscriberCount(streamData models.StreamData, rules []models.RuleResponse) error {

	serviceStreamData := models.StreamData{}
	for _, rule := range rules {
		if rule.Operation == utils.SUBSCRIPTION_COUNT {

			for _, serviceUUID := range streamData.ServiceList {
				serviceStreamData.UUID = serviceUUID
				// check if this service is subscribe for more than subscription threshold
				updateDb, err := d.IsServiceEligibleForTag(serviceStreamData, rule)
				if err != nil {
					return err
				}
				fmt.Println("IsServiceEligibleForTag :", updateDb)

				if updateDb {
					d.UpdateTagToService(streamData, rule)
				}
			}
		}
	}
	return nil
}
