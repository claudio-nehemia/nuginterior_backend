package middleware

import (
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

// RequirePermission checks if the current user's role has the specified permission.
// Permissions are cached in Redis via AuthService.GetPermissionsByRoleID.
func RequirePermission(authSvc service.AuthService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDVal, exists := c.Get("role_id")
		if !exists {
			helper.Forbidden(c, "Akses ditolak: role tidak ditemukan")
			c.Abort()
			return
		}

		roleID, ok := roleIDVal.(uint)
		if !ok || roleID == 0 {
			helper.Forbidden(c, "Akses ditolak: role tidak valid")
			c.Abort()
			return
		}

		perms, err := authSvc.GetPermissionsByRoleID(c.Request.Context(), roleID)
		if err != nil {
			helper.InternalError(c, "Gagal memuat permissions")
			c.Abort()
			return
		}

		for _, p := range perms {
			if p == permission {
				c.Next()
				return
			}
		}

		helper.Forbidden(c, "Akses ditolak: anda tidak memiliki permission '"+permission+"'")
		c.Abort()
	}
}
