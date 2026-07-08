package database

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"gorm.io/gorm"
)

// CompanyScope scopes GORM queries to the user's company_id.
// Super Admin of Company ID 1 can override this using the ContextKeyFilterCompanyID context.
func CompanyScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if ctx == nil {
			return db
		}

		companyID, ok := ctx.Value(constants.ContextKeyCompanyID).(uint)
		if !ok {
			return db
		}

		role, _ := ctx.Value(constants.ContextKeyUserRole).(string)

		// Super Admin of Company ID 1 has global filter privileges
		if companyID == 1 && role == "Super Admin" {
			if filterVal := ctx.Value(constants.ContextKeyFilterCompanyID); filterVal != nil {
				if filterID, ok := filterVal.(uint); ok && filterID > 0 {
					return db.Where("company_id = ?", filterID)
				}
				// If filter is explicitly set to 0 (all), do not apply any company_id filter
				return db
			}
			return db
		}

		// Everyone else is strictly scoped to their own company
		return db.Where("company_id = ?", companyID)
	}
}

// OrderScope scopes GORM queries to the orders belonging to the user's company_id.
func OrderScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if ctx == nil {
			return db
		}

		companyID, ok := ctx.Value(constants.ContextKeyCompanyID).(uint)
		if !ok {
			return db
		}

		role, _ := ctx.Value(constants.ContextKeyUserRole).(string)

		// Super Admin of Company ID 1 has global filter privileges
		if companyID == 1 && role == "Super Admin" {
			if filterVal := ctx.Value(constants.ContextKeyFilterCompanyID); filterVal != nil {
				if filterID, ok := filterVal.(uint); ok && filterID > 0 {
					return db.Where("order_id IN (SELECT id FROM orders WHERE company_id = ?)", filterID)
				}
				return db
			}
			return db
		}

		// Everyone else is strictly scoped to their own company's orders
		return db.Where("order_id IN (SELECT id FROM orders WHERE company_id = ?)", companyID)
	}
}

// GetContextCompanyID returns the target company_id for creation/updates based on context.
// It respects the active filter company_id if the user is a Super Admin of Company 1.
func GetContextCompanyID(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}

	companyID, ok := ctx.Value(constants.ContextKeyCompanyID).(uint)
	if !ok {
		return 0
	}

	role, _ := ctx.Value(constants.ContextKeyUserRole).(string)

	// If Super Admin of Company 1, check if they are filtering a specific company
	if companyID == 1 && role == "Super Admin" {
		if filterVal := ctx.Value(constants.ContextKeyFilterCompanyID); filterVal != nil {
			if filterID, ok := filterVal.(uint); ok && filterID > 0 {
				return filterID
			}
		}
	}

	return companyID
}

