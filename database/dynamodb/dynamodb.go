package dynamodb

import (
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

	fmt.Println("1CreateService : ServiceUUID : ", service.ServiceUUID)
	// if its a fresh entry
	if service.ServiceUUID == "" {
		service.ServiceUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		service.CreatedAt, service.UpdatedAt = datetime, datetime
		service.PK = utils.GetPartitionKey(utils.SERVICE, service.ServiceName, blank)
	}

	fmt.Println("2CreateService : service.PK : ", service.PK)

	av, err := dynamodbattribute.MarshalMap(service)
	if err != nil {
		return service, err
	}

	fmt.Println("3CreateService : service.PK : ", service.PK)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName.MDSTable),
	}

	fmt.Println("4CreateService : service.PK : ", d.tableName.MDSTable)
	fmt.Println("4CreateService : service.PK : ", input)

	_, err = d.db.PutItem(input)
	if err != nil {
		return service, err
	}

	fmt.Println("5CreateService : service.PK  :  ", input)

	return service, nil
}

func (d *Database) GetAllServices() ([]models.ServiceResponse, error) {

	services := []models.ServiceResponse{}
	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.SERVICE, blank, blank)

	keyCond := expression.Key(pkName).BeginsWith(pkPrefix)

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
	pk := utils.GetPartitionKey(utils.SERVICE, name, blank)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
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

func (d *Database) GetServiceByUUID(uuid string) (models.ServiceResponse, error) {

	services := []models.ServiceResponse{}
	// pkName := utils.GetPartitionKeyName(utils.SERVICE)
	// pk := utils.GetPartitionKey(utils.SERVICE, name)

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uuid": {
				S: aws.String(uuid),
			},
		},
		KeyConditionExpression: aws.String("uuid = :uuid"),
		TableName:              aws.String(d.tableName.MDSTable),
		IndexName:              aws.String("facility-code-index"),
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

func (d *Database) UpdateService(updatedService models.ServiceRequest) error {

	uuid := updatedService.ServiceUUID
	oldService, err := d.GetServiceByUUID(uuid)
	if err != nil {
		return err
	}

	oldServiceName := utils.GetPartitionKey(utils.SERVICE, oldService.ServiceName, blank)
	newServiceName := utils.GetPartitionKey(utils.SERVICE, updatedService.ServiceName, blank)

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
	updatedService.PK = utils.GetPartitionKey(utils.SERVICE, updatedService.ServiceName, blank)
	_, err = d.CreateService(updatedService)
	if err != nil {
		// TODO: restore old entry in case of error
		return err
	}

	return nil
}

func (d *Database) DeleteService(name string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.SERVICE, name, blank)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
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
		company.CompanyUUID = utils.GetUUID()
		datetime := utils.DateString("datetime")
		company.CreatedAt, company.UpdatedAt = datetime, datetime
		company.PK = utils.GetPartitionKey(utils.COMPANY, company.CompanyName, blank)
	}

	av, err := dynamodbattribute.MarshalMap(company)
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

func (d *Database) GetAllCompanies() ([]models.CompanyResponse, error) {

	companies := []models.CompanyResponse{}
	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.COMPANY, blank, blank)

	keyCond := expression.Key(pkName).BeginsWith(pkPrefix)

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

	// get latest service name
	for cindex, company := range companies {
		for sindex, srv := range company.ServiceList {
			if srv.ServiceUUID != "" {
				service, err := d.GetServiceByUUID(srv.ServiceUUID)
				if err == nil {
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
	pk := utils.GetPartitionKey(utils.COMPANY, name, blank)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
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

	// get latest service name
	for sindex, srv := range company.ServiceList {
		if srv.ServiceUUID != "" {
			service, err := d.GetServiceByUUID(srv.ServiceUUID)
			if err == nil {
				company.ServiceList[sindex].ServiceName = service.ServiceName
			}
		}
	}

	return company, nil
}

func (d *Database) GetCompanyByUUID(uuid string) (models.CompanyResponse, error) {

	companies := []models.CompanyResponse{}

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uuid": {
				S: aws.String(uuid),
			},
		},
		KeyConditionExpression: aws.String("uuid = :uuid"),
		TableName:              aws.String(d.tableName.MDSTable),
		IndexName:              aws.String("facility-code-index"),
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

func (d *Database) UpdateCompany(updatedCompany models.CompanyRequest) error {

	uuid := updatedCompany.CompanyUUID
	oldCompany, err := d.GetCompanyByUUID(uuid)
	if err != nil {
		return err
	}

	oldCompanyName := utils.GetPartitionKey(utils.COMPANY, oldCompany.CompanyName, blank)
	newCompanyName := utils.GetPartitionKey(utils.COMPANY, updatedCompany.CompanyName, blank)

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
	updatedCompany.PK = utils.GetPartitionKey(utils.COMPANY, updatedCompany.CompanyName, blank)
	err = d.CreateCompany(updatedCompany)
	if err != nil {
		// TODO: restore old entry in case of error
		return err
	}

	return nil
}

func (d *Database) DeleteCompany(name string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.COMPANY, name, blank)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
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
func (d *Database) CreateTag(tag models.TagCreateRequest) error {

	datetime := utils.DateString("datetime")
	tag.CreatedAt, tag.UpdatedAt = datetime, datetime
	tag.PK = utils.GetPartitionKey(utils.TAG, tag.Key, tag.Value)

	av, err := dynamodbattribute.MarshalMap(tag)
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

func (d *Database) GetAllTags() ([]models.TagListResponse, error) {

	tags := []models.TagResponse{}
	tagList := make([]models.TagListResponse, 0)

	pkName := utils.GetPartitionKeyName()
	pkPrefix := utils.GetPartitionKey(utils.TAG, blank, blank)

	keyCond := expression.Key(pkName).BeginsWith(pkPrefix)

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

	for _, tag := range tags {
		tagMap[tag.Key] = append(tagMap[tag.Key], tag.Value)
	}

	for key, value := range tagMap {
		temp := models.TagListResponse{
			Key:    key,
			Values: value,
		}
		tagList = append(tagList, temp)
	}

	return tagList
}

func (d *Database) DeleteTag(key string, value string) error {

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.TAG, key, value)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
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

func (d *Database) GetTag(key string) (models.TagListResponse, error) {

	tags := []models.TagResponse{}
	resultTag := []models.TagListResponse{}
	dummy := models.TagListResponse{}

	pkName := utils.GetPartitionKeyName()
	pk := utils.GetPartitionKey(utils.TAG, key, blank)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(pk),
			},
		},
		TableName: aws.String(d.tableName.MDSTable),
	}

	result, err := d.db.GetItem(input)
	if err != nil {
		return dummy, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &tags)
	if err != nil {
		return dummy, err
	}

	resp := createTagResponse(tags, resultTag)
	if len(resp) > 0 {
		return createTagResponse(tags, resultTag)[0], nil
	}

	return dummy, nil
}
