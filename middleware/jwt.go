package middleware

import (
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/serializer"
	"net/http"
	"time"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		//var data interface{}

		code = 200
		token := c.Request.Header.Get("token")
		if token == "" {
			code = -1 // invalid token error, to be set
		} else {
			claims, err := auth.ParseToken(token)
			if err != nil {
				code = -1 // token error code, to be set
			} else if time.Now().Unix() > claims.ExpiresAt {
				code = -1 // timeout error code, to be set
			}
		}
		if code != 200 {
			c.JSON(http.StatusUnauthorized, serializer.Response{
				Code: code,
				Msg: "token error",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

