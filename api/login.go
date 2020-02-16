package api

import (
	"fmt"
	"jxc/serializer"
	"jxc/service"

	"github.com/gin-gonic/gin"
)

// 登录
func Login(c *gin.Context) {
	// 获取 当前域名绑定了哪个公司
	/*	host := strings.Split(c.Request.Host, ":")[0]
		company, err := GetCompany(host)
		if err != nil {
			// 域名未注册或找不到这家公司
			c.JSON(200, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}*/

	var service service.UserLoginService
	fmt.Println(service.UserName)
	fmt.Println(service.Password)
	if err := c.ShouldBind(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		// userInfo, err := service.Login("", form.Username, form.Password)
		// if err != nil {
		// 	c.JSON(200, serializer.Response{
		// 		Code: -1,
		// 		Msg:  err.Error(),
		// 	})
		// 	return
		c.JSON(200, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
	}

	// 登录成功，把用户信息保存成session
	// session := sessions.Default(c)
	// session.Set("user_id", userInfo.UserId)
	// session.Set("username", userInfo.Username)
	// session.Set("com_id", userInfo.ComId)
	// session.Set("phone", userInfo.Phone)
	// session.Set("authority", userInfo.Authority)
	// session.Set("position", userInfo.Position)
	// session.Save()

	// c.JSON(200, serializer.Response{
	// 	Code: 0,
	// 	Msg:  "登录成功",
	// 	//Data: userInfo,
	// })

}

// ErrorResponse 返回错误消息
// func ErrorResponse(err error) serializer.Response {
// 	if ve, ok := err.(validator.ValidationErrors); ok {
// 		for _, e := range ve {
// 			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
// 			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
// 			return serializer.ParamErr(
// 				fmt.Sprintf("%s%s", field, tag),
// 				err,
// 			)
// 		}
// 	}
// 	if _, ok := err.(*json.UnmarshalTypeError); ok {
// 		return serializer.ParamErr("JSON类型不匹配", err)
// 	}

// 	return serializer.ParamErr("参数错误", err)
// }
