package database

import "github.com/auto-tagging-mds/database/models"

type Database interface {
	CreateService(models.ServiceRequest) (models.ServiceRequest, error)
	GetAllServices() ([]models.ServiceResponse, error)
	GetService(name string) (models.ServiceResponse, error)
	UpdateService(models.ServiceRequest, string) error
	DeleteService(name string) error

	CreateCompany(models.CompanyRequest) error
	GetAllCompanies() ([]models.CompanyResponse, error)
	GetCompany(name string) (models.CompanyResponse, error)
	UpdateCompany(models.CompanyRequest, string) error
	DeleteCompany(name string) error

	CreateTag(models.TagCreateRequest) (models.TagCreateRequest, error)
	GetAllTags() ([]models.TagListResponse, error)
	DeleteTag(key string, value string) error
	GetTag(key string) (models.TagListResponse, error)
	// UpdateTag(models.TagRequest) error
	// DeleteTag(name string) error

	
}
