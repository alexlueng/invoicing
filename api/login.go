package api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginService struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

// 登录
func Login(c *gin.Context) {


	var service LoginService
	var user models.User

	if err := c.ShouldBind(&service); err == nil {
		fmt.Println("username: ", service.Username)
		fmt.Println("password: ", service.Password)

		err = models.Client.Collection("users").FindOne(context.TODO(), bson.D{{"username", service.Username}}).Decode(&user)
		if err != nil {
			fmt.Println("Can't find user.")
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg: "Can't find user.",
			})
			return
		}

		if !user.CheckPassword(user.Password) {
			fmt.Println("Password error")
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg: "Password error",
			})
			return
		}

		token, err := auth.GenerateToken(user.Username, user.Password)
		if err != nil {
			fmt.Println("can't generate token.")
			return
		}
		fmt.Println("Generate token: ", token)
		//res := service.Login(c)
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg: "Login success",
			Data: token,
		})
	} else {
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
