package wxapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/api"
	"jxc/models"
	"jxc/models/wxapp"
	"jxc/serializer"
	"jxc/util"
	"net/http"
	"strings"
	"time"
)

// 微信小程序商城api
// 商品，分类等列表的返回

/*
	用户首次登录获得openid并存到数据库中，下次登录的时候从数据库中先检测用户是否已经存在openid,
	如果有，则认为用户已经登录
	获取用户的openid并保存到数据库中
	url: https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JSCODE&grant_type=authorization_code
*/

const (
	URL       = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	APPID     = "wxf0467c5997b35ffe"
	APPSECRET = "3a54adbe7988050115ee748d12172bd1"
)

type WechatLogin struct {
	JsCode string `json:"jscode"`
	OpenID string `json:"open_id"` // TODO：现在直接传一个open_id, 以后会传一个加密过的token
}

type WechatLoginResp struct {
	SessionKey string `json:"session_key"`
	OpenID     string `json:"openid"`
}

// 小程序用户登录前的检查接口
func CheckLogin(c *gin.Context) {
	var openID WechatLogin
	if err := c.ShouldBindJSON(&openID); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("miniapp_user")
	var miniappUser models.MiniAppUser
	err := collection.FindOne(context.TODO(), bson.D{{"open_id", openID.OpenID}}).Decode(&miniappUser)
	//TODO：目前是找到这个openid就认为他已经登录了，之后要根据session_key来判断这个用户是否已经过期，如果过期的话就要重新登录
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeMiniappUserNotLoginErr,
			Msg:  "user not login",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "user has been login",
	})
}

/*
		用户首次进入小程序，用户的openid没有在数据库中，则新建一个用户
	    如果是新用户进入小程序的话，提交code到后端，后端在数据库创建新的用户，此时的新用户数据只有少量数据的（如后端自己创建的id），所以我们需要将能获取到的值传给后端完善数据库。
		一、判断数据库中此用户是否有头像和名称，没有则需要授权用户数据（传头像和名称），将得到的头像和名称传到后端完善数据库，此时需要用到用户信息的授权，即wx.getUserInfo。
		二、判断数据库中此用户是否有手机号，没有则调用获取手机号授权，即getPhoneNumber。
		微信小程序中,每次向后台发送request默认都是一个全新的会话,如想要进行会话保持,可以在登录后返回sessionid保存,以后再向服务器发送请求时可以在请求头加上sessionid,来保证会话与上次会话一致
*/

// 小程序用户登录
func Login(c *gin.Context) {

	// 获取小程序端发送用户的jscode
	var jscode WechatLogin
	if err := c.ShouldBindJSON(&jscode); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	fmt.Println("jscode: ", jscode.JsCode)
	url := fmt.Sprintf(URL, APPID, APPSECRET, jscode.JsCode) // 向微信获取用户的openid
	fmt.Println("Wechat url: ", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "http request error",
		})
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	client := http.Client{}                                              //创建客户端
	resp, err := client.Do(request.WithContext(context.TODO()))          //发送请求
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "http response error",
		})
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.Log().Error("Get wechat auth error: ", err.Error())
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Get wechat auth error",
		})
		return
	}

	var wechatResp WechatLoginResp
	err = json.Unmarshal(data, &wechatResp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if len(wechatResp.OpenID) == 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Get open id error",
		})
		return
	}

	// 返回登录状态前端
	fmt.Println("response data: ", string(data))
	// 如果数据库里没有这个openid, 则新建一个，如果已经有了，则不创建
	collection := models.Client.Collection("miniapp_user")
	// 创建一个新用户
	var miniappUser models.MiniAppUser

	err = collection.FindOne(context.TODO(), bson.D{{"open_id", wechatResp.OpenID}}).Decode(&miniappUser)
	if err == nil && len(miniappUser.OpenID) > 0 {
		// 说明数据库里已经有了这个openid,不创建新用户
		util.Log().Info("Get miniapp user: ", miniappUser)

		// 返回cart_id
		collection = models.Client.Collection("user_cart")
		var cart wxapp.UserCart
		err = collection.FindOne(context.TODO(), bson.D{{"open_id", miniappUser.OpenID}}).Decode(&cart)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Find user cart error",
			})
			return
		}

		// 返回用户的购物车信息
		collection = models.Client.Collection("cart_item")
		var cartItems []wxapp.CartItem
		cur, err := collection.Find(context.TODO(), bson.D{{"open_id", miniappUser.OpenID}, {"is_delete", false}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Find user cart item error",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var res wxapp.CartItem
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode cart item error",
				})
				return
			}
			cartItems = append(cartItems, res)
		}



		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "miniapp user has already exist",
			Data: map[string]interface{}{
				"open_id": wechatResp.OpenID,
				"cart": cartItems,
				"cart_id" : cart.CartID,
				"user_id" : miniappUser.UserID,
			},
		})
		return
	}

	miniappUser.UserID = api.GetLastID("miniapp_user")
	miniappUser.OpenID = wechatResp.OpenID
	miniappUser.SessionKey = wechatResp.SessionKey
	miniappUser.CreateAt = time.Now().Unix()
	miniappUser.Verify = 0

	_, err = collection.InsertOne(context.TODO(), miniappUser)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't create new user",
		})
		return
	}
	var cart wxapp.UserCart
	if !CheckMiniappUserCart(0, 0, miniappUser.OpenID) {
		cart, err = CreateCart(0, 0, miniappUser.OpenID)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't create new user cart",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Create new miniapp user",
		Data: map[string]interface{}{
			"open_id": wechatResp.OpenID,
			"cart": nil,
			"cart_id" : cart.CartID,
		},
	})
}

// 小程序传过来的userInfo，完善用户信息
type MiniappUserInfoService struct {
	OpenID   string             `json:"openId"`
	UserInfo models.MiniAppUser `json:"userInfo"`
}

func GetUserInfo(c *gin.Context) {

	var srv MiniappUserInfoService
	if err := c.ShouldBindJSON(&srv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("miniapp_user")
	var user models.MiniAppUser
	err := collection.FindOne(context.TODO(), bson.D{{"open_id", srv.OpenID}}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "No user error",
		})
		return
	}

	user.AvatarURL = srv.UserInfo.AvatarURL
	user.City = srv.UserInfo.City
	user.Gender = srv.UserInfo.Gender
	user.Language = srv.UserInfo.Language
	user.NickName = srv.UserInfo.NickName
	user.Province = srv.UserInfo.Province

	_, err = collection.UpdateOne(context.TODO(), bson.D{{"open_id", user.OpenID}}, bson.M{"$set": user})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Update user error",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update user",
	})
	//_ = json.Unmarshal(data, &d)
}

type VerifyMaterialSerive struct {
	LicenseURL  string `json:"license_url"`
	CompanyType int64  `json:"company_type"`
	Company     string `json:"company"`
	LicenseNo   string `json:"license_no"`
	Contact     string `json:"contact"`
	Telephone   string `json:"telephone"`
	VerifyCode  string `json:"verify_code"`
	Address     string `json:"address"`
}

// 用户上传营业执照等信息
// TODO:验证码是否正确
func UserVerify(c *gin.Context) {

	var userVerifySrv VerifyMaterialSerive
	if err := c.ShouldBindJSON(&userVerifySrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	userVerify := models.UserVerifyMaterial{
		MaterialID: api.GetLastID("material"),
		// TODO: 传一个用户ID，也可以通过token来解出来
		UserID:          0,
		BusinessLicense: userVerifySrv.LicenseURL,
		LicenseNo:       userVerifySrv.LicenseNo,
		CompanyType:     userVerifySrv.CompanyType,
		CompanyName:     userVerifySrv.Company,
		Contact:         userVerifySrv.Contact,
		Telephone:       userVerifySrv.Telephone,
		Address:         userVerifySrv.Address,
		VerifyCode:      userVerifySrv.VerifyCode,
	}

	collection := models.Client.Collection("material")
	_, err := collection.InsertOne(context.TODO(), userVerify)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Verify material upload success",
	})
}

// 发送手机验证码
func SendVerifyCode(c *gin.Context) {}

// 页面上展示待审核的页面
func ListVerifyMaterial(c *gin.Context) {}

// 首页优选商品列表
func PreferredProductList(c *gin.Context) {

	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	util.Log().Info("get com id: ", com.ComId)

	var products []models.Product
	collection := models.Client.Collection("product")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["preferred"] = true
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find preferred product error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode preferred product error",
			})
			return
		}
		products = append(products, res)
	}

	// 返回商品图片
	productImages := make(map[int64][]models.Image)
	collection = models.Client.Collection("image")
	for _, product := range products {

		cur, err := collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}, {"product_id", product.ProductID}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find images",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var image models.Image
			if err := cur.Decode(&image); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't decode images",
				})
				return
			}
			productImages[product.ProductID] = append(productImages[product.ProductID], image)
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product list",
		Data: map[string]interface{}{
			"products": products,
			"images":   productImages,
		},
	})
}

// 首页推荐商品列表
func RecommandProductList(c *gin.Context) {

	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	util.Log().Info("get com id: ", com.ComId)

	var products []models.Product
	collection := models.Client.Collection("product")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["recommand"] = true
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find recommand product error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode recommand product error",
			})
			return
		}
		products = append(products, res)
	}

	// 返回商品图片
	productImages := make(map[int64][]models.Image)
	collection = models.Client.Collection("image")
	for _, product := range products {

		cur, err := collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}, {"product_id", product.ProductID}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find images",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var image models.Image
			if err := cur.Decode(&image); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't decode images",
				})
				return
			}
			productImages[product.ProductID] = append(productImages[product.ProductID], image)
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product list",
		Data: map[string]interface{}{
			"products": products,
			"images":   productImages,
		},
	})
}

func CategoryList(c *gin.Context) {

	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	var categories []models.Category
	collection := models.Client.Collection("category")

	cur, err := collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find category error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Category
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode category error",
			})
			return
		}
		categories = append(categories, res)
	}

	var images []models.CategoryImage
	collection = models.Client.Collection("category_image")
	cur, err = collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find category images",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.CategoryImage
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode image error",
			})
			return
		}
		images = append(images, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Category list",
		Data: map[string]interface{}{
			"categories": categories,
			"images":     images,
		},
	})
}

