package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
)

type ReqLogin struct {
	Username string `json:"username" form:"username"`
	Phone    string `json:"phone" form:"phone"`
	Password string `json:"password" form:"password"`
}

// 登录
func Login(c *gin.Context) {
	// 获取请求的域名，可以得知所属公司

	domain := c.Request.Header.Get("Origin")
	com, err := models.GetComIDAndModuleByDomain(domain[len("http://"):])
		if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	var user models.User
	var req ReqLogin

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	filter := bson.M{}
	// 用户名登录
	if req.Username != "" {
		filter["username"] = req.Username
	}
	// 手机号登录
	if req.Phone != "" {
		filter["phone"] = req.Phone
	}

	// TODO：判断是否超级管理员
	collection := models.Client.Collection("company")
	var admin models.Company
	err = collection.FindOne(context.TODO(), bson.D{{"phone", req.Phone}, {"com_id", com.ComId}}).Decode(&admin)

	if err != nil { // 不是超级管理员
		collection = models.Client.Collection("users")

		filter["com_id"] = com.ComId

		err = collection.FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该用户不存在！",
			})
			return
		}
		//pwd, _ := util.PasswordBcrypt("123456")

		if !util.PasswordVerify(req.Password, user.Password) {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "密码错误！",
			})
			return
		}
		token, _ := auth.GenerateToken(user.Username, user.UserID, com.ComId, false)
		// 获取这个用户的所有权限路由节点id，在根据节点id获取所有路由
		filter = bson.M{}
		auth_note := models.AuthNote{}
		urls := []string{}
		filter["auth_id"] = bson.M{"$in": user.Authority} // TODO: 要判断user.Authority里面是否有值，不然程序会报错
		// 新建完用户之后马上跳到权限设置页面
		cur, _ := models.Client.Collection("auth_note").Find(context.TODO(), filter)
		//defaultUrl := []string{"/api/v1/units"}
		for cur.Next(context.TODO()) {
			err = cur.Decode(&auth_note)
			if err != nil {
				continue
			}
			for _, val := range auth_note.Urls {
				urls = append(urls, val)
			}
		}
		user.Urls = urls
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Login success",
			Data: map[string]interface{}{
				"username": user.Username,
				"phone":    user.Phone,
				"position": user.Position,
				"token":    token,
				"urls":     user.Urls,
			},
		})
		return
	}

	// 如果是超级管理员
	if GenMD5Password(req.Password) != admin.Password {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "管理员密码错误！",
		})
		return
	}

	token, _ := auth.GenerateToken(admin.Admin, admin.ComId, com.ComId, true)
	urls := []string{}
	cur, _ := models.Client.Collection("auth_note").Find(context.TODO(), bson.D{})
	//defaultUrl := []string{"/api/v1/units"}
	auth_note := models.AuthNote{}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&auth_note)
		if err != nil {
			continue
		}
		for _, val := range auth_note.Urls {
			urls = append(urls, val)
		}
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Login success",
		Data: map[string]interface{}{
			"username": admin.Admin,
			"phone":    admin.Telephone,
			"position": "admin",
			"token":    token,
			"urls":     urls,
		},
	})
}

// md5
func GenMD5Password(passwd string) string {
	digest := md5.Sum([]byte(passwd))
	return hex.EncodeToString(digest[:])
}
