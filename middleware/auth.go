package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"jxc/util"
	"net/http"
	"strings"
	"time"
)

// 权限判断流程
// headers 中 Access-Token 为空，只能访问登录接口
// 不为空则进行解析
// 解析出错则踢出
// 当前域名是否为该公司所有，否则域名错误踢出
// 查询数据库，获取当前用户的所有权限路由
// 判断当前路由是否有权限
// 其他函数里获取com_id,user_id,解析token即可

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Access-Token")
		url := c.Request.URL.String()
		if url == "/api/v1/login" || url == "/add_superadmin" {
			c.Next()
			return
		}
		urlArr := strings.Split(url,"?")
		url = urlArr[0]

		claims, err := auth.ParseToken(token)
		fmt.Println("jwt claims: ", claims)
		//api.SmartPrint(&claims)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -2,
				Msg:  "token失效，请重新登录",
			})
			c.Abort()
			return
		}

		if claims.ComId == 0 || claims.ExpiresAt == 0 || claims.UserId == 0 {
			c.Abort()
			return
		}
		if time.Now().Unix() > claims.ExpiresAt {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -2,
				Msg:  "token失效，请重新登录",
			})
			c.Abort()
			return
		}
		// 验证当前域名是否注册

		domain := c.Request.Header.Get("Origin")
		fmt.Println("访问域名： ", domain[7:])
		com, err := models.GetComIDAndModuleByDomain(domain[7:])
		//moduleID, _ := strconv.Atoi(com.ModuleId)

		if err != nil || models.THIS_MODULE != com.ModuleId {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			c.Abort()
			return
		}
		// 用了别的公司的域名
		if com.ComId != claims.ComId {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "域名错误",
			})
			c.Abort()
			return
		}

		if claims.Admin { // 是超级管理员
			c.Next()
			return
		}

		// 查询当前用户的信息
		user, err := service.FindOneUser(claims.UserId, claims.ComId)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "没有这个用户",
			})
			c.Abort()
			return
		}
		if !util.InArray(user.Urls, url) {
			fmt.Println("url:",url)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -2,
				Msg:  "没有这项权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CurrentUser 获取登录用户
func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// uid := session.Get("user_id")
		// if uid != nil {
		// 	user, err := models.GetUser(uid)
		// 	if err == nil {
		// 		c.Set("user", &user)
		// 	}
		// }
		c.Next()
	}
}

// AuthRequired 需要登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get("user"); user != nil {
			if _, ok := user.(*models.User); ok {
				c.Next()
				return
			}
		}

		c.JSON(200, serializer.CheckLogin())
		c.Abort()
	}
}
