package service

import (
	"context"
	"fmt"
	"jxc/models"
	"jxc/serializer"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	UserName string `form:"username" json:"username" binding:"required,min=5,max=30"`
	Password string `form:"password" json:"password" binding:"required"`
}

// Login handle the user login
func (u *UserLoginService) Login(c *gin.Context) serializer.Response {
	fmt.Println(c.GetRawData())
	var user models.User

	err := models.Client.Collection("users").FindOne(context.TODO(), bson.D{{"username", u.UserName}}).Decode(&user)
	if err != nil {
		return serializer.ParamErr("账号或密码错误", err)
	}
	if !user.CheckPassword(u.Password) {
		return serializer.ParamErr("密码错误", nil)
	}

	u.setSession(c, user)

	return serializer.BuildUserResponse(user)
}

// setSession 设置session
func (service *UserLoginService) setSession(c *gin.Context, user models.User) {
	s := sessions.Default(c)
	s.Clear()
	s.Set("user_id", user.UserID)
	s.Save()
}

// add a comment
