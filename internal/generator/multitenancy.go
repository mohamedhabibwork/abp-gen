package generator

import (
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
)

// MultiTenancyHelper provides utilities for multi-tenancy features
type MultiTenancyHelper struct{}

// NewMultiTenancyHelper creates a new multi-tenancy helper
func NewMultiTenancyHelper() *MultiTenancyHelper {
	return &MultiTenancyHelper{}
}

// IsEnabled checks if multi-tenancy is enabled
func (h *MultiTenancyHelper) IsEnabled(sch *schema.Schema) bool {
	return sch.Solution.MultiTenancy != nil && sch.Solution.MultiTenancy.Enabled
}

// GetStrategy returns the multi-tenancy strategy
func (h *MultiTenancyHelper) GetStrategy(sch *schema.Schema) string {
	if !h.IsEnabled(sch) {
		return "none"
	}
	return sch.Solution.MultiTenancy.Strategy
}

// ShouldAddTenantFilter checks if tenant filter should be applied
func (h *MultiTenancyHelper) ShouldAddTenantFilter(sch *schema.Schema, entity *schema.Entity) bool {
	if !h.IsEnabled(sch) {
		return false
	}

	// Don't add tenant filter for value objects
	if entity.EntityType == "ValueObject" {
		return false
	}

	// Add tenant filter if data isolation is enabled
	return sch.Solution.MultiTenancy.EnableDataIsolation
}

// GetTenantIdProperty returns the tenant ID property name
func (h *MultiTenancyHelper) GetTenantIdProperty(sch *schema.Schema) string {
	if !h.IsEnabled(sch) || sch.Solution.MultiTenancy.TenantIdProperty == "" {
		return "TenantId"
	}
	return sch.Solution.MultiTenancy.TenantIdProperty
}

// NeedsMultiTenancyAttribute checks if entity needs [MultiTenant] attribute
func (h *MultiTenancyHelper) NeedsMultiTenancyAttribute(sch *schema.Schema, entity *schema.Entity) bool {
	return h.ShouldAddTenantFilter(sch, entity)
}

// GetConnectionStringStrategy returns connection string strategy for multi-tenancy
func (h *MultiTenancyHelper) GetConnectionStringStrategy(sch *schema.Schema) string {
	if !h.IsEnabled(sch) {
		return "default"
	}

	strategy := h.GetStrategy(sch)
	switch strategy {
	case "tenant-per-db":
		return "per-tenant"
	case "tenant-per-schema":
		return "per-schema"
	default:
		return "shared"
	}
}

// ShouldAllowCrossTenant checks if cross-tenant queries are allowed
func (h *MultiTenancyHelper) ShouldAllowCrossTenant(sch *schema.Schema) bool {
	if !h.IsEnabled(sch) {
		return false
	}
	return sch.Solution.MultiTenancy.EnableCrossTenant
}

// BuildMultiTenancyConfig builds configuration data for templates
func (h *MultiTenancyHelper) BuildMultiTenancyConfig(sch *schema.Schema, entity *schema.Entity) map[string]interface{} {
	return map[string]interface{}{
		"Enabled":            h.IsEnabled(sch),
		"Strategy":           h.GetStrategy(sch),
		"TenantIdProperty":   h.GetTenantIdProperty(sch),
		"NeedsFilter":        h.ShouldAddTenantFilter(sch, entity),
		"NeedsAttribute":     h.NeedsMultiTenancyAttribute(sch, entity),
		"ConnectionStrategy": h.GetConnectionStringStrategy(sch),
		"AllowCrossTenant":   h.ShouldAllowCrossTenant(sch),
	}
}
