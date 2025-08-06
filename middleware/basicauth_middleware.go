package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

func BasicAuthMiddleware() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		os.Getenv("BASICAUTH_ADMIN_USERNAME"): os.Getenv("BASICAUTH_ADMIN_PASSWORD"),
	})
}
