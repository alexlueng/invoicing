package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strconv"
	//"gopkg.in/go-playground/validator.v9"
)

//允许同名的客户
const ENABLESAMECUSTOMER = false

// http://localhost:3000/api/v1/customer/list?page=1&size=10&name=0&level=0 返回json 结果
func ListCustomers(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.CustReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}
	// 设置分页的默认值
	page, size := SetDefaultPageAndSize(req.Page, req.Size)
	// 设置排序主键
	orderFields := []string{"customer_id", "com_id", "customer_name", "level", "payment", "payamount", "receiver", "address", "phone"}
	option := SetPaginationAndOrder(req.OrdF, orderFields, req.Ord, page, size)

	filter := bson.M{}
	if req.Name != "" {
		filter["name"] = bson.M{"$regex": req.Name}
	}
	if req.Receiver != "" {
		filter["receiver"] = bson.M{"$regex": req.Receiver}
	}
	level, _ := strconv.Atoi(req.Level)
	if level > 0 {
		filter["level"] = bson.M{"$eq": level}
	}
	payment, _ := strconv.Atoi(req.Payment)
	if payment != 0 {
		filter["payment"] = bson.M{"$eq": req.Payment}
	}
	if req.Address != "" {
		filter["address"] = bson.M{"$regex": req.Address}
	}
	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}

	filter["com_id"] = claims.ComId

	var customers []models.Customer
	customer :=  models.Customer{}

	customers, err = customer.FindAll(filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 返回客户等级
	var levels []models.Level
	levelResult, err := models.SelectLevelByComID(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	for _, res := range levelResult.Level {
		levels = append(levels, res)
	}

	//查询的总数
	total, _ := customer.Total(filter)

	// 返回查询到的总数，总页数
	resData := models.ResponseCustomerData{}
	resData.Customers = customers
	resData.Levels = levels
	resData.Total = total
	resData.Pages = total/size + 1
	resData.Size = size
	resData.CurrentPage = page

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get customers",
		Data: resData,
	})

}

func AddCustomer(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)
	customer := models.Customer{}
	_ = json.Unmarshal(data, &customer)

	customer.ComID = claims.ComId

	if !ENABLESAMECUSTOMER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		if customer.CheckExist() {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "该客户已经存在",
			})
			return
		}
	}
	customer.ID = GetLastID("customer")

	err := customer.Insert()
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "创建客户失败",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Customer create succeeded",
		Data: customer,
	})
}

// UpdateCustomer update an exist record
func UpdateCustomer(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	updateCus := models.Customer{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateCus)

	updateCus.ComID = claims.ComId

	//应当使用validator 来检验参数是否正确

	//检查是否重名
	if !updateCus.UpdateCheck() {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Customer name exist",
		})
		return
	}
	if err := updateCus.Update(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Customer update failed",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer update succeeded",
	})
}

type DeleteCustomerService struct {
	ID int64 `json:"customer_id"`
}

// DeleteCustomer delete an exist record
func DeleteCustomer(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var d DeleteCustomerService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	customer := models.Customer{
		ComID: claims.ComId,
		ID: d.ID,
	}
	if err := customer.Delete(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Customer delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer delete succeeded",
	})
}

type CustomerCount struct {
	NameField string
	Count     int
}

