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

	CreateTag(models.TagCreateRequest) error
	GetAllTags() ([]models.TagListResponse, error)
	DeleteTag(key string, value string) error
	GetTag(key string) (models.TagListResponse, error)
	// UpdateTag(models.TagRequest) error
	// DeleteTag(name string) error

	// UpdateHotelList(*models.OTAHotelSearchRSEnvelope, string, string) error
	// UpdateRoomPlan(*models.OTAHotelAvailRoomRSEnvelope, string, string, string) error
	// RoomSync(*models.OTAHotelAvailRoomRSEnvelope, string, string, string) error
	// PlanSync(*models.OTAHotelAvailRoomRSEnvelope, string, string, string) error
	// PlanCancellationPolicySync(*models.OTAHotelAvailRoomRSEnvelopeCP1, string, string, string) error
	// OptionSync(*models.OTAHotelAvailRoomRSEnvelope, string, string, string) error
	// InventorySync(*models.OTAHotelAvailRoomRSEnvelope, string, string, string) error
	// PriceSync(*models.OTAHotelAvailRoomRSEnvelope, string, string, string, int) error
	// HotelList() (*[]*models.SiteControllerUser, error)
	// GetHotelByID(string) (*models.SiteControllerUser, error)
	// GetReservablePeriod(string) (int64, error)
	// GetTLIDForPlanRoom(string, string, string) (string, string, error)
	// GetHotelName(string) (string, string, error)
	// GetBookingInfoTest(string, string) (be.Booking, error)
	// InsertTLBookingID(*models.OTAHotelResRSEnvelope, string) error
	// InsertTLBookingIDError(string, string, []string) error
	// GetTLBookingID(string, string) (string, error)
	// UpdateTLBookingID(string, string, string) error
	// RoomUpdateTimeUpdate(string, string) error
	// PlanUpdateTimeUpdate(string, string) error
	// InventoryUpdateTimeUpdate(string, string) error
	// PriceUpdateTimeUpdate(string, string) error
	// OptionUpdateTimeUpdate(string, string) error
	// GetRoomUpdateTime(string) (models.UpdateStatus, error)
	// GetPlanUpdateTime(string) (models.UpdateStatus, error)
	// GetInventoryUpdateTime(string) (models.UpdateStatus, error)
	// GetPriceUpdateTime(string) (models.UpdateStatus, error)
	// GetOptionUpdateTime(string) (models.UpdateStatus, error)
	// AddSyncRequest(string, string) error
	// TLBEOptionIDMapList(string) (map[string]string, error)
	// TLBERoomIDMapList(string) (*[]*models.RoomIDMapList, error)
	// TLBEPlanIDMapList(string) (*[]*models.PlanIDMapList, error)
	// GetPlanInfoList(string) (*[]*be.Plan, error)
}
