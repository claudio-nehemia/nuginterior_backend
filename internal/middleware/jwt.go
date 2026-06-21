package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth validates the JWT access token from the Authorization header.
// On success it sets user_id, email, role_id, and jti in the gin context.
func JWTAuth(cfg *config.Config, authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helper.Unauthorized(c, "Token diperlukan")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			helper.Unauthorized(c, "Format token tidak valid")
			c.Abort()
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			helper.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			helper.Unauthorized(c, "Token claims tidak valid")
			c.Abort()
			return
		}

		// Check token type
		if claims["type"] != "access" {
			helper.Unauthorized(c, "Token bukan access token")
			c.Abort()
			return
		}

		// Check blacklist
		jti, _ := claims["jti"].(string)
		if authSvc.IsTokenBlacklisted(c.Request.Context(), jti) {
			helper.Unauthorized(c, "Token sudah di-revoke")
			c.Abort()
			return
		}

		// Set context values
		userIDFloat, _ := claims["user_id"].(float64)
		c.Set("user_id", uint(userIDFloat))

		email, _ := claims["email"].(string)
		c.Set("email", email)

		companyIDFloat, _ := claims["company_id"].(float64)
		companyID := uint(companyIDFloat)
		c.Set("company_id", companyID)

		roleName, _ := claims["role_name"].(string)
		c.Set("role_name", roleName)

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, constants.ContextKeyUserEmail, email)
		reqCtx = context.WithValue(reqCtx, constants.ContextKeyCompanyID, companyID)
		reqCtx = context.WithValue(reqCtx, constants.ContextKeyUserRole, roleName)

		// Check for X-Company-Filter header if user is Super Admin of Company 1
		if companyID == 1 && roleName == "Super Admin" {
			filterHeader := c.GetHeader("X-Company-Filter")
			if filterHeader != "" && filterHeader != "all" {
				var filterID uint
				if _, err := fmt.Sscanf(filterHeader, "%d", &filterID); err == nil && filterID > 0 {
					reqCtx = context.WithValue(reqCtx, constants.ContextKeyFilterCompanyID, filterID)
					c.Set("filter_company_id", filterID)
				}
			} else if filterHeader == "all" {
				reqCtx = context.WithValue(reqCtx, constants.ContextKeyFilterCompanyID, uint(0))
				c.Set("filter_company_id", uint(0))
			}
		}

		c.Request = c.Request.WithContext(reqCtx)

		c.Set("jti", jti)

		if roleID, ok := claims["role_id"].(float64); ok {
			c.Set("role_id", uint(roleID))
		}

		// Store expiry for logout
		if exp, ok := claims["exp"].(float64); ok {
			c.Set("token_exp", int64(exp))
		}

		c.Next()
	}
}
