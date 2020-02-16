package api

import (
	"context"
	"encoding/json"
	"fmt"
	"jxc/models"
	"jxc/serializer"
	"strings"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//允许同名的客户
const ENABLESAMECUSTOMER = false

// http://localhost:3000/api/v1/customer/list?page=1&size=10&name=0&level=0 返回json 结果
// 默认展示前20条数据，第1页，以升序的方式
func ListCustomers(c *gin.Context) {

	// 根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req models.CustReq
	var customers []models.Customer

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	// 设置分页的默认值
	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderField := []string{"customer_id", "com_id", "customer_name", "level", "payment", "payamount", "receiver", "address", "phone"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "customer_id"
	}
	// 设置排序顺序 desc asc
	order := 1
	fmt.Println("order: ", req.Ord)
	if req.Ord == "desc" {
		order = -1
		req.Ord = "desc"
	} else {
		order = 1
		req.Ord = "asc"
	}

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))

	//1从小到大,-1从大到小
	option.SetSort(bson.D{{req.OrdF, order}})

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
	filter["com_id"] = com.ComId

	fmt.Println("filter: ", filter)

	// 当前请求页面的数据
	cur, err := models.Client.Collection("customer").Find(ctx, filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}

	for cur.Next(context.TODO()) {
		var result models.Customer
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding customer: ", err)
			return
		}
		customers = append(customers, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("customer").Find(ctx, filter)
	for cur.Next(context.TODO()) {
		total++
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseCustomerData{}
	resData.Customers = customers
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customers",
		Data: resData,
	})

}

// AddCustomer a customer and save into mongodb
func AddCustomer(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}
	data, _ := ioutil.ReadAll(c.Request.Body)
	customer := models.Customer{}

	_ = json.Unmarshal(data, &customer)
	collection := models.Client.Collection("customer")
	result := models.Customer{}

	if !ENABLESAMECUSTOMER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		filter := bson.M{}
		filter["com_id"] = com.ComId
		filter["name"] = customer.Name
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Name != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该客户已经存在",
			})
			return
		}
	}
	customer.ID = int64(getLastCustomerID())

	insertResult, err := collection.InsertOne(context.TODO(), customer)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer create succeeded",
	})

}

// UpdateCustomer update an exist record
func UpdateCustomer(c *gin.Context) {

	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	updateCus := models.Customer{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateCus)

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的customer_id相等,如果有不相等的，则返回
	// 如果只有相等的customer_id, 则允许修改
	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["name"] = updateCus.Name
	collection := models.Client.Collection("customer")

	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempRes models.Customer
		err := cur.Decode(&tempRes)
		if err != nil {
			fmt.Println("error found decoding customer: ", err)
			return
		}
		if tempRes.ID != updateCus.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的客户名已经存在",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["customer_id"] = updateCus.ID
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"name": updateCus.Name,
			"receiver": updateCus.Receiver,
			"receiver_phone": updateCus.Phone,
			"receiver_address": updateCus.Address,
			"payment": updateCus.Payment,
			"level": updateCus.Level}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新失败",
		})
		return
	}
	fmt.Println("Update result: ", result	)
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

	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	if ( (err != nil) || (models.THIS_MODULE != com.ModuleId) ){
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	var d DeleteCustomerService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["customer_id"] = d.ID
	collection := models.Client.Collection("customer")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除用户失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
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

