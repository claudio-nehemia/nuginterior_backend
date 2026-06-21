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
