package api

import (
	"github.com/gin-gonic/gin"
	"jxc/models"
	"jxc/serializer"

	"net/http"

	"jxc/auth"
)

type AuthData struct {
	Username string	`valid:"Required; MaxSize(50)"`
	Password string `valid:"Required; MaxSize(50)"`
}

func GetAuth(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	//valid := validator.New()
	//a := AuthData{Username:username, Password: password}
	ok, _ := true, true

	data := make(map[string]interface{})
	code := -1 // invalid params code

	if ok {
		isExist := models.CheckAuth(username, password)
		if isExist {
			token, err := auth.GenerateToken(username, password)
			if err != nil {
				code = -1 // error auth token
			} else {
				data["token"] = token
				code = 200
			}
		} else {
			code = -1 // error auth
		}
	} else {
		//for _, err := valid.Errors {
		//	log.Println(err.Key, err.Message)
		//}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: code,
		Msg: "Auth failed",
	})
}
