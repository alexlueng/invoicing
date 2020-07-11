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
		util.Log().Info("request url: ", url)
		if url == "/api/v1/login" || url == "/add_superadmin" || url == "/api/v1/supplier/mobile/login" || strings.Index(url, "wxapp") != -1 || strings.Index(url, "wx") != -1 {
			c.Next()
			return
		}
		urlArr := strings.Split(url,"?")
		url = urlArr[0]

		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeTokenErr,
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
				Code: serializer.CodeTokenErr,
				Msg:  "token失效，请重新登录",
			})
			c.Abort()
			return
		}
		// 验证当前域名是否注册

		domain := c.Request.Header.Get("Origin")
		com, err := models.GetComIDAndModuleByDomain(domain[len("http://"):])

		if err != nil || models.THIS_MODULE != com.ModuleId {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			c.Abort()
			return
		}
		// 用了别的公司的域名
		if com.ComId != claims.ComId {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "域名错误",
			})
			c.Abort()
			return
		}

		if claims.Admin { // 是超级管理员
			c.Next()
			return
		}

		// 判断是供应商平台登录还是进销存平台
		if strings.Index(url, "mobile") != -1 {
			var supplier models.Supplier
			supplier.ComID = claims.ComId
			_, err := supplier.FindByID(claims.UserId)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "没有这个供应商",
				})
				c.Abort()
				return
			}
		} else {
			// 查询当前用户的信息
			_, err = service.FindOneUser(claims.UserId, claims.ComId)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "没有这个用户",
				})
				c.Abort()
				return
			}
		}

		// TODO: 这里关闭了用户权限的验证，上线的时候要把它开启
		// TODO: 建议这里改成casbin
		//if !util.InArray(user.Urls, url) {
		if false {
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
