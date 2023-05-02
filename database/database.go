package database

import "github.com/auto-tagging-mds/database/models"

type Database interface {
	CreateService(models.ServiceRequest) (models.ServiceRequest, error)
	GetAllServices() ([]models.ServiceResponse, error)
	GetService(name string) (models.ServiceResponse, error)
	UpdateService(models.ServiceRequest, string) error
	DeleteService(name string) error

	CreateCompany(models.CompanyRequest) (models.CompanyRequest, error)
	GetAllCompanies() ([]models.CompanyResponse, error)
	GetCompany(name string) (models.CompanyResponse, error)
	UpdateCompany(models.CompanyRequest, string) error
	DeleteCompany(name string) error

	CreateTag(models.TagCreateRequest) (models.TagCreateRequest, error)
	GetAllTags() ([]models.TagListResponse, error)
	DeleteTag(key string, value string) error
	GetTag(key string, value string) (models.TagListResponse, error)

	CreateRule(models.RuleRequest) (models.RuleRequest, error)
	GetAllRules() ([]models.RuleResponse, error)
	GetRule(ruleUUID string) (models.RuleResponse, error)
	UpdateRule(models.RuleRequest, string) error
	DeleteRule(ruleUUID string) error

	AttachTagWithService(service models.StreamData, rules []models.RuleResponse) error
	ProcessRuleForServices(models.StreamData, []models.ServiceResponse) error
	UpdateServiceTagForSubscriberCount(streamData models.StreamData, rules []models.RuleResponse) error
}
