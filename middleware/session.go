package middleware

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"jxc/auth"
)

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	store := cookie.NewStore([]byte(secret))
	//Also set Secure: true if using SSL, you should though
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"})
	return sessions.Sessions("gin-session", store)
}


// Try to get com_id from middleware
func GetComIDAndModuleID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据域名得到com_id
		token := c.GetHeader("Access-Token")
		claims, _ := auth.ParseToken(token)
		fmt.Println("ComID: ", claims.ComId)
		c.Set("com_id", claims.ComId)
		c.Next()
	}
}