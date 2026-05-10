package httpd

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Recovery Gin panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[Recovery] panic:", err)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}