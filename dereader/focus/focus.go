// Package focus implements the FinOps Cost and Usage Specification (FOCUS)
package focus

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type Resource struct {
	OrgID      string
	SpaceID    string
	InstanceID string

	OrgName   string
	SpaceName string

	SvcPlanName  string
	SvcOfferName string
}

type Reader interface {
	IsMetered() bool

	GetResource() *Resource
	GetResourceID() string
}

type Spec struct {
	Resource *Resource

	AvailabilityZone           string              `json:"availabilityZone,omitempty"`
	BilledCost                 decimal.NullDecimal `json:"billedCost"`
	BillingAccountID           string              `json:"billingAccountId,omitempty"`
	BillingAccountName         string              `json:"billingAccountName,omitempty"`
	BillingCurrency            string              `json:"billingCurrency"`
	BillingPeriodEnd           time.Time           `json:"billingPeriodEnd"`
	BillingPeriodStart         time.Time           `json:"billingPeriodStart"`
	CapacityReservationID      string              `json:"capacityReservationId,omitempty"`
	CapacityReservationStatus  string              `json:"capacityReservationStatus,omitempty"`
	ChargeCategory             string              `json:"chargeCategory"`
	ChargeClass                string              `json:"chargeClass,omitempty"`
	ChargeDescription          string              `json:"chargeDescription"`
	ChargeFrequency            string              `json:"chargeFrequency,omitempty"`
	ChargePeriodEnd            time.Time           `json:"chargePeriodEnd"`
	ChargePeriodStart          time.Time           `json:"chargePeriodStart"`
	CommitmentDiscountCategory string              `json:"commitmentDiscountCategory,omitempty"`
	CommitmentDiscountID       string              `json:"commitmentDiscoutId,omitempty"`
	CommitmentDiscountType     string              `json:"commitmentDiscountType,omitempty"`
	CommitmentDiscountStatus   string              `json:"commitmentDiscountStatus,omitempty"`
	CommitmentDiscountName     string              `json:"commitmentDiscountName,omitempty"`
	CommitmentDiscountQuantity decimal.NullDecimal `json:"commitmentDiscountQuantity"`
	CommitmentDiscountUnit     string              `json:"commitmentDiscountUnit,omitempty"`
	ConsumedQuantity           decimal.NullDecimal `json:"consumedQuantity"`
	ConsumedUnit               string              `json:"consumedUnit"`
	ContractedCost             decimal.NullDecimal `json:"contractedCost"`
	ContractedUnitCost         decimal.NullDecimal `json:"contractedUnitCost"`
	EffectiveCost              decimal.NullDecimal `json:"effectiveCost"`
	InvoiceIssuerName          string              `json:"invoiceIssuerName"`
	ListCost                   decimal.NullDecimal `json:"listCost"`
	ListUnitPrice              decimal.NullDecimal `json:"listUnitPrice"`
	PricingCategory            string              `json:"pricingCategory,omitempty"`
	PricingQuantity            decimal.NullDecimal `json:"pricingQuantity"`
	PricingUnit                string              `json:"pricingUnit,omitempty"`
	ProviderName               string              `json:"providerName,omitempty"`
	PublisherName              string              `json:"publisherName,omitempty"`
	RegionID                   string              `json:"regionId,omitempty"`
	RegionName                 string              `json:"regionName,omitempty"`
	ResourceID                 string              `json:"resourceId,omitempty"`
	ResourceName               string              `json:"resourceName,omitempty"`
	ResourceType               string              `json:"resourceType,omitempty"`
	ServiceCategory            string              `json:"serviceCategory"`
	ServiceName                string              `json:"serviceName,omitempty"`
	ServiceSubcategory         string              `json:"serviceSubcategory,omitempty"`
	SkuID                      string              `json:"skuId,omitempty"`
	SkuMeter                   string              `json:"skuMeter,omitempty"`
	SkuPriceDetails            Tags                `json:"skuPriceDetails,omitempty"`
	SkuPriceID                 string              `json:"skuPriceId,omitempty"`
	SubAccountID               string              `json:"subAccountId,omitempty"`
	SubAccountName             string              `json:"subAccountName,omitempty"`
	Tags                       Tags                `json:"tags,omitempty"`
}

type Tags map[string]string

func (t *Tags) UnmarshalCSV(s string) error {
	if s == "" {
		return nil
	}

	err := json.Unmarshal([]byte(s), t)
	if err != nil {
		return err
	}

	return nil
}
