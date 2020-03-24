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
	"net/http"
	"strconv"
	//"gopkg.in/go-playground/validator.v9"
)

//允许同名的客户
const ENABLESAMECUSTOMER = false

// http://localhost:3000/api/v1/customer/list?page=1&size=10&name=0&level=0 返回json 结果
// 默认展示前20条数据，第1页，以升序的方式
func ListCustomers(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)


	var req models.CustReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	// 设置分页的默认值
	page, size := SetDefaultPageAndSize(req.Page, req.Size)
	// 设置排序主键
	orderFields := []string{"customer_id", "com_id", "customer_name", "level", "payment", "payamount", "receiver", "address", "phone"}
	option := SetPaginationAndOrder(req.OrdF, orderFields, req.Ord, page, size)


	// TODO:这些搜索条件是否可以从这个函数中提取出来
	//IdMin,IdMax
	if req.IdMin > req.IdMax {
		t := req.IdMax
		req.IdMax = req.IdMin
		req.IdMin = t
	}
	filter := bson.M{}
	if (req.IdMin == req.IdMax) && (req.IdMin != 0) {
		//filter["id"] = bson.M{"$gte":0}
		filter["id"] = bson.M{"$eq": req.IdMin}
	} else {
		if req.IdMin > 0 {
			filter["id"] = bson.M{"$gte": req.IdMin}
		}
		if req.IdMax > 0 {
			filter["id"] = bson.M{"$lt": req.IdMax}
		}
	}
	// Reciever string `form:"reciever"` //模糊搜索
	if req.Name != "" {
		filter["name"] = bson.M{"$regex": req.Name}
	}
	// Reciever string `form:"reciever"` //模糊搜索
	if req.Receiver != "" {
		fmt.Println("receiver: ", req.Receiver)
		filter["receiver"] = bson.M{"$regex": req.Receiver}
	}
	// Level    int    `form:"level"`
	level, _ := strconv.Atoi(req.Level)
	if level > 0 {
		filter["level"] = bson.M{"$eq": level}
	}
	//Payment  int    `form:"payment"`
	payment, _ := strconv.Atoi(req.Payment)
	if payment != 0 {
		filter["payment"] = bson.M{"$eq": req.Payment}
	}
	// Address  string `form:"address"`  //模糊搜索
	if req.Address != "" {
		filter["address"] = bson.M{"$regex": req.Address}
	}
	// Phone    string `form:"phone"`    //模糊搜索
	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}

	// 每个查询都要带着com_id去查
	//com_id, _ := strconv.Atoi(com.ComId)
	filter["com_id"] = claims.ComId

	var customers []models.Customer
	customer :=  models.Customer{}

	customers, err = customer.FindAll(filter, option)
	if err != nil {
		fmt.Println("error found decoding customer: ", err)
		return
	}

	//查询的总数
	total, _ := customer.Total(filter)

	// 返回查询到的总数，总页数
	resData := models.ResponseCustomerData{}
	resData.Customers = customers
	resData.Total = total
	resData.Pages = total/size + 1
	resData.Size = size
	resData.CurrentPage = page

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customers",
		Data: resData,
	})

}

// AddCustomer a customer and save into mongodb
func AddCustomer(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	data, _ := ioutil.ReadAll(c.Request.Body)
	customer := models.Customer{}
	_ = json.Unmarshal(data, &customer)

	customer.ComID = claims.ComId


	//validate := validator.New()
	//validate.RegisterValidation("my-validate", customFunc)
	//err = validate.Struct(customer)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(err)
	//SmartPrint(customer)
	//collection := models.Client.Collection("customer")
	//result := models.Customer{}

	if !ENABLESAMECUSTOMER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		if customer.CheckExist() {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该客户已经存在",
			})
			return
		}
	}
	customer.ID = int64(getLastCustomerID())

	err := customer.Insert()
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer create succeeded",
	})
}

// UpdateCustomer update an exist record
func UpdateCustomer(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

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
	fmt.Println("ComID: ", claims.ComId)

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

// 因mongodb不允许自增方法，所以要生成新增客户的id
// 这是极度不安全的代码，因为本程序是分布式的，本程序可能放在多台服务器上同时运行的。
// 需要在交付之前修改正确
func getLastCustomerID() int {
	var cc CustomerCount
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", "customer"}}).Decode(&cc)
	if err != nil {
		fmt.Println("can't get customerID")
		return 0
	}
	collection.UpdateOne(context.TODO(), bson.M{"name": "customer"}, bson.M{"$set": bson.M{"count": cc.Count + 1}})
	fmt.Println("customer count: ", cc.Count)
	return cc.Count + 1
}

