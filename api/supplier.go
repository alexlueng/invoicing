package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)
//允许同名的供应商
const ENABLESAMESUPPLIER = false


func ListSuppliers(c *gin.Context) {
	// 根据域名得到com_id
	// 用postman调试的时候需要注释
	//com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	//if err != nil || models.THIS_MODULE != int(com.ModuleId) {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "域名错误",
	//	})
	//	return
	//}

	var req models.SupplierReq
	var suppliers []models.Supplier

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderField := []string{"supplier_id", "com_id"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "supplier_id"
	}
	// 设置排序顺序 desc asc
	order := 1
	fmt.Println("order: ", req.Ord)
	if req.Ord == "desc" {
		order = -1
		//req.Ord = "desc"
	} else {
		order = 1
		//req.Ord = "asc"
	}

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))

	option.SetSort(bson.D{{req.OrdF, order}})

	// 页面搜索
	filter := bson.M{}
	//IdMin,IdMax
	if req.IdMin > req.IdMax {
		t := req.IdMax
		req.IdMax = req.IdMin
		req.IdMin = t
	}
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
	//ID int64 `json:"supplier_id" form:"supplier_id"`
	if req.ID > 0 {
		filter["supplier_id"] = bson.M{"$eq": req.ID}
	}
	//Phone string `json:"phone" form:"phone"`
	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}
	//SupplierName string `json:"supplier_name" form:"supplier_name"`
	if req.SupplierName != "" {
		filter["supplier_name"] = bson.M{"$regex": req.SupplierName}
	}
	//Level int64 `json:"level" form:"level"`
	if req.Level > 0 {
		filter["level"] = bson.M{"$regex": req.Level}
	}
	//Payment string `json:"payment" form:"payment"`
	if req.Payment != "" {
		filter["payment"] = bson.M{"$regex": req.Payment}
	}


	// 每个查询都要带着com_id
	//filter["com_id"] = com.ComId
	filter["com_id"] = 1
	// all conditions are set then start searching
	collection := models.Client.Collection(("supplier"))
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Supplier
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding recording: ", err)
			return
		}
		suppliers = append(suppliers, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("supplier").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	resData := models.ResponseSupplierData{}
	resData.Suppliers = suppliers
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get suppliers",
		Data: resData,
	})
}
func AddSuppliers(c *gin.Context) {
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
	supplier := models.Supplier{}

	_ = json.Unmarshal(data, &supplier)
	collection := models.Client.Collection("supplier")
	result := models.Supplier{}



	if !ENABLESAMESUPPLIER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		filter := bson.M{}
		//filter["com_id"] = com.ComId
		filter["com_id"] = 1
		filter["supplier_name"] = supplier.SupplierName
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.SupplierName != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该供应商已经存在",
			})
			return
		}
	}
	supplier.ID = int64(getLastID("supplier"))
	supplier.ComID = 1
	SmartPrint(supplier)
	insertResult, err := collection.InsertOne(context.TODO(), supplier)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Supplier create succeeded",
	})
}
func UpdateSuppliers(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	updateSupplier := models.Supplier{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateSupplier)

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的supplier_id相等,如果有不相等的，则返回
	// 如果只有相等的supplier_id, 则允许修改
	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["supplier_name"] = updateSupplier.SupplierName
	collection := models.Client.Collection("supplier")

	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempRes models.Supplier
		err := cur.Decode(&updateSupplier)
		if err != nil {
			fmt.Println("error found decoding supplier: ", err)
			return
		}
		if tempRes.ID != updateSupplier.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的供应商已经存在",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["customer_id"] = updateSupplier.ID
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"supplier_name": updateSupplier.SupplierName,
			"contacts": updateSupplier.Contacts,
			"phone": updateSupplier.Phone,
			"address": updateSupplier.Address,
			"payment": updateSupplier.Payment,
			"level": updateSupplier.Level}})
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
		Msg:  "Supplier update succeeded",
	})
}

type DeleteSupplierService struct {
	ID int64 `json:"supplier_id"`
}

func DeleteSuppliers(c *gin.Context) {

	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	if ( (err != nil) || (models.THIS_MODULE != com.ModuleId) ){
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	var d DeleteSupplierService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["supplier_id"] = d.ID
	collection := models.Client.Collection("supplier")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除供应商失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Supplier delete succeeded",
	})
}