package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/shopspring/decimal"
)

type ComputeAddress struct {
	Address                string
	Region                 string
	AddressType            string
	Purpose                string
	InstancePurchaseOption string

	// usage args
	AddressUsageType *string `infracost_usage:"address_type"`
}

var ComputeAddressUsageSchema = []*schema.UsageItem{
	{Key: "address_type", ValueType: schema.String, DefaultValue: ""},
}

func (r *ComputeAddress) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeAddress) BuildResource() *schema.Resource {
	addressType := r.AddressType
	isFreePurpose := r.Purpose != "" && strings.ToLower(r.Purpose) != "gce_endpoint"

	if strings.ToLower(addressType) == "internal" || isFreePurpose {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: ComputeAddressUsageSchema,
		}
	}

	costComponents := []*schema.CostComponent{}

	usageType, err := r.detectUsageType()
	if err != "" {
		log.Warnf(err)
	}

	switch usageType {
	case "standard_vm":
		costComponents = append(costComponents, r.standardVMComputeAddress(1))
	case "preemptible_vm":
		costComponents = append(costComponents, r.preemptibleVMComputeAddress(1))
	case "unused":
		costComponents = append(costComponents, r.unusedVMComputeAddress(1))
	default:
		costComponents = append(costComponents, r.unusedVMComputeAddress(0))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    ComputeAddressUsageSchema,
	}
}

func (r *ComputeAddress) standardVMComputeAddress(count int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if count > 0 {
		quantity = decimalPtr(decimal.NewFromInt(count))
	}

	return &schema.CostComponent{
		Name:           "IP address (standard VM)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Standard VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("744"),
		},
	}
}

func (r *ComputeAddress) preemptibleVMComputeAddress(count int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if count > 0 {
		quantity = decimalPtr(decimal.NewFromInt(count))
	}

	return &schema.CostComponent{
		Name:           "IP address (preemptible VM)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Spot Preemptible VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

func (r *ComputeAddress) unusedVMComputeAddress(count int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if count > 0 {
		quantity = decimalPtr(decimal.NewFromInt(count))
	}

	return &schema.CostComponent{
		Name:           "IP address (unused)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Static Ip Charge")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

func (r *ComputeAddress) detectUsageType() (string, string) {
	validTypes := []string{"standard_vm", "preemptible_vm", "unused"}

	usageType := ""

	if r.AddressUsageType == nil {
		switch r.InstancePurchaseOption {
		case "on_demand":
			return "standard_vm", ""
		case "preemptible":
			return "preemptible_vm", ""
		}

		return usageType, ""
	}

	usageType = strings.ToLower(*r.AddressUsageType)

	if !contains(validTypes, usageType) {
		return "", fmt.Sprintf("Invalid address_type, ignoring. Expected: standard_vm, preemptible_vm, unused. Got: %s", *r.AddressUsageType)
	}

	return usageType, ""
}
