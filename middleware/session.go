package middleware

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"jxc/models"
	"strings"
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
		com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
		//fmt.Println("com_id: ", com.ComId)
		//fmt.Println("module_id: ", com.ModuleId)
		if err != nil || com.ModuleId != models.THIS_MODULE {
			fmt.Println("error found while getting com id: ", err)
			c.Abort()
			return
		}
		c.Set("com_id", com.ComId)
		c.Next()
	}
}