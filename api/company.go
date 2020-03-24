package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
)

//
type Company struct {
	ComID     string              `json:"com_id" form:"com_id"`
	ComName   string              `json:"com_name" form:"com_name"`
	Delivery  string              `json:"delivery" form:"delivery"`
	Domain    string              `json:"domain" form:"domain"`
	Units     string              `json:"units" form:"units"`
	Developer string              `json:"developer" form:"developer"`
	Domains   []models.DomainData `json:"domains" form:"domains"`
}

// 接收更新公司信息数据字段
type ComRequest struct {
	ComName  string               `json:"com_name" form:"com_name"`
	Delivery []ComDeliveryRequest `json:"delivery" form:"delivery[]"` // 快递方式
	Units    []string             `json:"units" form:"units[]"`       // 商品量词
	//Developer string   `json:"developer" form:"developer"` // 开发商名称
	//Domains   []string `json:"domains" form:"domains[]"`   // 域名
	//Module  string      `json:"module"  form:"module"`    //平台名称
	Payment             []string `json:"payment" form:"payment[]"`                           //结算方式
	Position            []string `form:"position[]" json:"position"`                         //职务
	DefaultProfitMargin int64    `json:"default_profit_margin" form:"default_profit_margin"` //默认利润率
}

// 接收到的 Delivery 字段格式
type ComDeliveryRequest struct {
	DeliveryCom    string `json:"delivery_com" form:"delivery_com"`       // 配送公司
	DeliveryPerson string `json:"delivery_person" form:"delivery_person"` // 配送员
	Phone          string `json:"phone" form:"phone"`                     // 配送员电话
	Config         string `json:"config" form:"config"`                   // 配置参数
}

// 接收Payment 字段格式
type ComPaymentRequest struct {
	PaymentId   int64  `json:"payment_id" form:"payment_id"`     // 支付方式id
	PaymentName string `json:"payment_name" form:"payment_name"` // 支付方式名称
	Days        int64  `json:"days" form:"days"`                 // 天数
}

// 返回公司详情数据格式
type ComInfoResponse struct {
	ComName             string          `json:"com_name" form:"com_name"`
	ExpirationDate      int64           `json:"expiration_date" form:"expiration_date"`             // 到期时间
	Delivery            interface{}     `json:"delivery" form:"delivery"`                           // 快递方式
	Units               interface{}     `json:"units" form:"units"`                                 // 商品量词
	Payment             interface{}     `json:"payment"`                                            // 支付方式
	Developer           string          `json:"developer" form:"developer"`                         // 开发商名称
	Domains             []models.Domain `json:"domains" form:"domains"`                             // 域名
	Module              string          `json:"module"  form:"module"`                              //平台名称
	Position            interface{}     `form:"position[]" json:"position"`                         //职务
	DefaultProfitMargin int64           `json:"default_profit_margin" form:"default_profit_margin"` //默认利润率
}

// 获取所有配送方式
func AllCompanies(c *gin.Context) {

	var companies []Company

	com1 := Company{
		ComID:     "1",
		ComName:   "huazhi01",
		Delivery:  "shunfeng",
		Domain:    "www.huazhi01.com",
		Units:     "pounds",
		Developer: "alex",
	}
	companies = append(companies, com1)
	com2 := Company{
		ComID:     "2",
		ComName:   "huazhi02",
		Delivery:  "yunda",
		Domain:    "www.huazhi02.com",
		Units:     "pounds",
		Developer: "bob",
	}
	companies = append(companies, com2)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Hello",
		Data: companies,
	})
}

// 获取公司信息
func CompanyDetail(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 指定数据集
	comCollection := models.Client.Collection("company")
	domainCollection := models.Client.Collection("domain")
	deliveryCollection := models.Client.Collection("delivery")

	var company models.Company
	var domains models.Domain
	var comInfoResponse ComInfoResponse
	var delivery models.Delivery
	var deliverys []models.Delivery
	filter := bson.M{}
	filter["com_id"] = claims.ComId

	// 查找公司相应的信息
	err := comCollection.FindOne(context.TODO(), filter).Decode(&company)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能找到公司的信息！",
		})
		return
	}
	// 获取公司配送信息
	cur, err := deliveryCollection.Find(context.TODO(), bson.M{"comid": claims.ComId})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能找到公司的信息！",
		})
		return
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&delivery)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "未能找到公司的信息！",
			})
			return
		}
		deliverys = append(deliverys, delivery)
	}
	// 获取公司支付信息

	comInfoResponse.ComName = company.ComName
	comInfoResponse.Delivery = deliverys //配送方式
	comInfoResponse.ExpirationDate = company.ExpirationDate
	comInfoResponse.Payment = company.Payment
	comInfoResponse.Developer = company.Developer
	comInfoResponse.Units = company.Units
	comInfoResponse.Position = company.Position
	comInfoResponse.DefaultProfitMargin = company.DefaultProfitMargin

	filter = bson.M{}
	filter["comid"] = claims.ComId

	// 找到公司下配置的所有域名
	cur, err = domainCollection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		err = cur.Decode(&domains)
		if err != nil {
			fmt.Println("error found decoding company: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "未能找到公司的信息！",
			})
			return
		}
		comInfoResponse.Domains = append(comInfoResponse.Domains, domains)
	}

	// 返回数据
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: comInfoResponse,
		Msg:  "",
	})

}

// 更新公司信息
func UpdateCompany(c *gin.Context) {
	// 获取请求的域名，可以得知所属公司
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ComRequest
	var deliverys []interface{}

	// 处理公司信息
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &req)
	if err != nil {
		fmt.Println("err found while decoding into company: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新公司信息失败！",
		})
		return
	}

	// 指定数据集
	comCollection := models.Client.Collection("company")
	deliveryCollection := models.Client.Collection("delivery")

	// 计量单位、支付方式、职务 去重
	req.Units = util.RemoveRepeatedElement(req.Units)
	req.Payment = util.RemoveRepeatedElement(req.Payment)
	req.Position = util.RemoveRepeatedElement(req.Position)

	// 整理配送方式数据
	for _, val := range req.Delivery {
		deliverys = append(deliverys, models.Delivery{
			ComId:          claims.ComId,
			DeliveryCom:    val.DeliveryCom,
			DeliveryPerson: val.DeliveryPerson,
			Phone:          val.Phone,
			Config:         val.Config,
		})
	}

	// 整理支付方式数据

	updateCom := bson.M{
		"units":                 req.Units,
		"comname":               req.ComName,
		"payment":               req.Payment,
		"delivery":              req.Delivery,
		"position":              req.Position,
		"default_profit_margin": req.DefaultProfitMargin,
	}

	// 更新公司信息
	_, err = comCollection.UpdateOne(context.TODO(), bson.M{"com_id": claims.ComId}, bson.M{
		"$set": updateCom,
	})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新公司信息失败！",
		})
		return
	}

	// 更新配送方式数据
	deliveryCollection.DeleteMany(context.TODO(), bson.M{"comid": claims.ComId})
	_, err = deliveryCollection.InsertMany(context.TODO(), deliverys)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新公司信息失败！",
		})
		return
	}

	// 返回公司信息
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: nil,
		Msg:  "更新公司信息成功！",
	})

}

// 获取配送方式
func DeliveryList(c *gin.Context) {
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	// 指定数据集
	collection := models.Client.Collection("delivery")
	delivery := models.Delivery{}
	deliverys := []models.Delivery{}
	filter := bson.M{}
	filter["comid"] = claims.ComId
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取配送方式失败！",
		})
		return
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&delivery)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "获取配送方式失败！",
			})
			return
		}
		deliverys = append(deliverys, delivery)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: deliverys,
		Msg:  "获取配送方式成功！",
	})
}

// 获取商品量词
func UnitsList(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 指定数据集
	collection := models.Client.Collection("company")
	company := models.Company{}
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	err := collection.FindOne(context.TODO(), filter).Decode(&company)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取商品量词失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: company.Units,
		Msg:  "获取商品量词成功！",
	})
}

// 获取结算方式
func PaymentList(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 指定数据集
	collection := models.Client.Collection("company")
	company := models.Company{}
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	err := collection.FindOne(context.TODO(), filter).Decode(&company)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取结算方式失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: company.Payment,
		Msg:  "获取结算方式成功！",
	})
}

// 获取默认利润率
func DefaultProfitMargin(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 指定数据集
	collection := models.Client.Collection("company")
	company := models.Company{}
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	err := collection.FindOne(context.TODO(), filter).Decode(&company)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取结算方式失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: map[string]int64{"default_profit_margin": company.DefaultProfitMargin},
		Msg:  "获取默认利润率成功！",
	})
}

// Get content:
// {"id":30,
// "module_id":1,
// "module_name":"进销存",
// "user_id":6,
// "admin":"alex",
// "telephone":"13888888888",
// "com_id":2,
// "company_name":"三国进销存",
// "domains":["4.vip.jxc.weqi.exechina.com"],
// "price":999,"create_at":1584762307,
// "last_paid_at":0,
// "expire_at":1585367107,
// "using":true,
// "is_try":true}

type NewInstance struct {
	ID          int64    `json:"id"`
	Admin       string   `json:"admin"`
	Password    string   `json:"password"`
	Telephone   string   `json:"telephone"`
	CompanyName string   `json:"company_name"`
	Domains     []string `json:"domains"`
	Using       bool     `json:"using"`
	ModuleID    int64    `json:"module_id"`
}

func AddSuperAdmin(c *gin.Context) {
	//content, _ := c.GetRawData()
	//fmt.Println("Get content: ", string(content))

	var req NewInstance
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}

	// TODO：添加超级管理员，修改域名表，Company表

	collection := models.Client.Collection("domain")
	for _, item := range req.Domains {
		var domain models.Domain
		domain.Domain = item
		domain.ComId = req.ID
		domain.ModuleId = req.ModuleID
		domain.Status = true
		insertResult, err := collection.InsertOne(context.TODO(), domain)
		if err != nil {
			fmt.Println("Can't insert domain： ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "Can't insert domain",
			})
			return
		}
		fmt.Println("Insert result: ", insertResult.InsertedID)
	}

	collection = models.Client.Collection("company")
	var defaultSuperadmin models.Company
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", 1}}).Decode(&defaultSuperadmin)
	if err != nil {
		fmt.Println("Can't find superadmin: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't find super user",
		})
		return
	}

	var admin models.Company
	admin.ComId = req.ID
	admin.ComName = req.CompanyName
	admin.Admin = req.Admin
	admin.Telephone = req.Telephone
	admin.Password = req.Password
	admin.Units = defaultSuperadmin.Units
	admin.Payment = defaultSuperadmin.Payment
	admin.Position = defaultSuperadmin.Position
	insertResult, err := collection.InsertOne(context.TODO(), admin)
	if err != nil {
		fmt.Println("Can't insert super admin: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't insert domain",
		})
		return
	}
	fmt.Println("Insert result: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Create new module succeeeeeeeeeed",
	})
}

// {"user_id":14,
// "username":"小李飞刀",
// "com_id":8,
// "password":"e10adc3949ba59abbe56e057f20f883e",
// "level":1,
// "telephone":"13999999999",
// "create_at":1584797806,
// "open_id":""}

type UpdateAdmin struct {
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
}

func UpdateAdminPasswd(c *gin.Context) {
	//content, _ := c.GetRawData()
	//fmt.Println("Get content: ", string(content))
	var updateAdmin UpdateAdmin
	if err := c.ShouldBindJSON(&updateAdmin); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}

	SmartPrint(updateAdmin)

	collection := models.Client.Collection("company")
	updateResult, err := collection.UpdateMany(context.TODO(), bson.D{{"telephone", updateAdmin.Telephone}}, bson.M{"$set": bson.M{"password": updateAdmin.Password}})
	if err != nil {
		fmt.Println("Can't update admin password: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}
	fmt.Println("update result: ", updateResult.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update admin password success",
	})
}

type DomainStatus struct {
	ComID  int64 `json:"com_id"`
	Status bool  `json:"status"`
}

func ChangeDomainStatus(c *gin.Context) {
	//content, _ := c.GetRawData()
	//fmt.Println("Get content: ", string(content))
	var d DomainStatus
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}

	SmartPrint(d)

	collection := models.Client.Collection("domain")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"comid", d.ComID}}, bson.M{"$set": bson.M{"status": d.Status}})
	if err != nil {
		fmt.Println("Update error: ", err)
		return
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update domain status success",
	})
}
