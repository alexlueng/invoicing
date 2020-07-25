package api

import (
	"encoding/json"
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
	ComName             string   `json:"com_name" form:"com_name"`
	Delivery            []int64  `json:"delivery" form:"delivery[]"`                         // 快递方式
	UnDelivery          []int64  `json:"un_delivery" form:"delivery[]"`                      // 不用的快递方式
	Units               []string `json:"units" form:"units[]"`                               // 商品量词
	Payment             []string `json:"payment" form:"payment[]"`                           //结算方式
	Position            []string `form:"position[]" json:"position"`                         //职务
	DefaultProfitMargin float64  `json:"default_profit_margin" form:"default_profit_margin"` //默认利润率
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
	CreateAt            int64           `json:"create_at" form:"create_at"`
	ExpireAt            int64           `json:"expire_at" form:"expire_at"`
	Delivery            interface{}     `json:"delivery" form:"delivery"`                           // 快递方式
	Units               interface{}     `json:"units" form:"units"`                                 // 商品量词
	Payment             interface{}     `json:"payment"`                                            // 支付方式
	Developer           string          `json:"developer" form:"developer"`                         // 开发商名称
	Domains             []models.Domain `json:"domains" form:"domains"`                             // 域名
	Module              string          `json:"module"  form:"module"`                              //平台名称
	Position            interface{}     `form:"position[]" json:"position"`                         //职务
	DefaultProfitMargin float64         `json:"default_profit_margin" form:"default_profit_margin"` //默认利润率
	Telephone           string          `json:"telephone" form:"telephone"`                         //超级管理员电话
	QRCodeURL           string          `json:"qrcode_url" form:"qrcode_url"`
	SupplierPlatformURL string          `json:"supplier_platform_url"` // 供应商平台
	ComID               int64           `json:"com_id" form:"com_id"`
	//ExpirationDate      int64           `json:"expiration_date" form:"expiration_date"`             // 到期时间
}

// 获取公司信息
func CompanyDetail(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var company *models.Company
	var comInfoResponse ComInfoResponse
	var deliverys []models.Delivery

	// 查找公司相应的信息
	company, err := models.SelectCompanyByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "未能找到公司的信息！",
		})
		return
	}

	// 获取公司配送信息
	deliveryRs, err := models.SelectDeliveryByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "未能找到公司的信息！",
		})
		return
	}
	for _, delivery := range deliveryRs.Delivery {
		deliverys = append(deliverys, delivery)
	}

	// 获取公司支付信息
	comInfoResponse.ComName = company.ComName
	comInfoResponse.Delivery = deliverys //配送方式
	comInfoResponse.Payment = company.Payment
	comInfoResponse.Developer = company.Developer
	comInfoResponse.Units = company.Units
	comInfoResponse.Position = company.Position
	comInfoResponse.DefaultProfitMargin = company.DefaultProfitMargin
	comInfoResponse.Telephone = company.Telephone
	comInfoResponse.CreateAt = company.CreateAt
	comInfoResponse.ExpireAt = company.ExpireAt
	comInfoResponse.QRCodeURL = "http://jxc.weqi.exechina.com/#/model-detail"
	comInfoResponse.ComID = company.ComId

	// 找到公司下配置的所有域名
	cur, err := models.SelectDomainByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "域名未注册",
		})
		return
	}
	for _, domains := range cur.Domain {
		comInfoResponse.Domains = append(comInfoResponse.Domains, domains)
	}

	comInfoResponse.SupplierPlatformURL = comInfoResponse.Domains[0].Domain + "/supplier"

	// 返回数据
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: comInfoResponse,
		Msg:  "",
	})

}

// 更新公司信息
func UpdateCompany(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ComRequest

	// 处理公司信息
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新公司信息失败！",
		})
		return
	}

	// 计量单位、支付方式、职务 去重
	req.Units = util.RemoveRepeatedElement(req.Units)
	req.Payment = util.RemoveRepeatedElement(req.Payment)
	req.Position = util.RemoveRepeatedElement(req.Position)

	//更新配送方式
	if len(req.Delivery) > 0 {
		err := models.UpdateDeliveryIsUsingFlag(claims.ComId, req.Delivery, true);
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "更新公司信息失败！",
			})
			return
		}
	}
	if len(req.UnDelivery) > 0 {
		err := models.UpdateDeliveryIsUsingFlag(claims.ComId, req.UnDelivery, false);
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "更新公司信息失败！",
			})
			return
		}
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
	err = models.UpdateCompanyByComID(claims.ComId, updateCom);
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "更新公司信息失败！",
		})
		return
	}
	// 返回公司信息
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: nil,
		Msg:  "更新公司信息成功！",
	})

}

// 获取配送方式
func DeliveryList(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 指定数据集
	deliverys := []models.Delivery{}

	rs, err := models.SelectDeliveryByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "获取配送方式失败！",
		})
		return
	}

	for _, delivery := range rs.Delivery {
		deliverys = append(deliverys, delivery)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: deliverys,
		Msg:  "获取配送方式成功！",
	})
}

// 获取商品量词
func UnitsList(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	company, err := models.SelectCompanyByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "获取商品量词失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: company.Units,
		Msg:  "获取商品量词成功！",
	})
}

// 获取结算方式
func PaymentList(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	company, err := models.SelectCompanyByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "获取结算方式失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: company.Payment,
		Msg:  "获取结算方式成功！",
	})
}

// 获取默认利润率
func DefaultProfitMargin(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	company, err := models.SelectCompanyByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "获取结算方式失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: map[string]float64{"default_profit_margin": company.DefaultProfitMargin},
		Msg:  "获取默认利润率成功！",
	})
}

type NewInstance struct {
	ID          int64    `json:"id"`
	Admin       string   `json:"admin"`
	Password    string   `json:"password"`
	Telephone   string   `json:"telephone"`
	CompanyName string   `json:"company_name"`
	Domains     []string `json:"domains"`
	Using       bool     `json:"using"`
	ModuleID    int64    `json:"module_id"`
	Developer   string   `json:"developer"`
	CreateAt    int64    `json:"create_at"`
	ExpireAt    int64    `json:"expire_at"`
}

func AddSuperAdmin(c *gin.Context) {

	var req NewInstance
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get data",
		})
		return
	}

	// 添加超级管理员，修改域名表，Company表
	// TODO:需要修改这个生成方式
	for _, item := range req.Domains {
		var domain models.Domain
		domain.Domain = item
		domain.ComId = req.ID
		domain.ModuleId = req.ModuleID
		domain.Status = true
		err := domain.Add()
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert domain",
			})
			return
		}
	}
	defaultSuperadmin, err := models.SelectCompanyByComID(1)
	if err != nil {
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
	admin.Developer = req.Developer
	admin.ExpireAt = req.ExpireAt
	admin.CreateAt = req.CreateAt

	admin.Units = defaultSuperadmin.Units
	admin.Payment = defaultSuperadmin.Payment
	admin.Position = defaultSuperadmin.Position

	err = admin.Add()
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't insert superuser",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Create new module succeed",
	})
}

type UpdateAdmin struct {
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
}

func UpdateAdminPasswd(c *gin.Context) {

	var updateAdmin UpdateAdmin
	if err := c.ShouldBindJSON(&updateAdmin); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}

	err := models.UpdateAdminPwdByTelPhone(updateAdmin.Telephone, updateAdmin.Password)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}
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
	var d DomainStatus
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}
	err := models.UpdateDomainStatusByComID(d.ComID, d.Status)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't update domain data",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update domain status success",
	})
}

type UpdateExipreService struct {
	ComID    int64 `json:"com_id"`
	ExpireAt int64 `json:"expire_at"`
}

// 更新模块过期时间
// TODO: 如果服务已经过期，则将该域名停用
func UpdateExpireTime(c *gin.Context) {
	var u UpdateExipreService
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}
	err := models.UpdateCompanyExpireTime(u.ComID, u.ExpireAt)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Can't get data",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update expire time success",
	})
}
