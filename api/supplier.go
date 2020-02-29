package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req models.SupplierReq

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	page, size := SetDefaultPageAndSize(req.Page, req.Size)

	//// 设置排序主键
	orderFields := []string{"supplier_id", "com_id"}
	option := SetPaginationAndOrder(req.OrdF, orderFields, req.Ord, page, size)

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
	if req.Contacts != "" {
		filter["contacts"] = bson.M{"$regex": req.Contacts}
	}

	if req.Name != "" {
		filter["supplier_name"] = bson.M{"$regex": req.Name}
	}

	// 每个查询都要带着com_id
	filter["com_id"] = com.ComId

	// all conditions are set then start searching
	var suppliers []models.Supplier
	supplier :=  models.Supplier{}

	suppliers, err = supplier.FindAll(filter, option)
	if err != nil {
		fmt.Println("error found decoding supplierS: ", err)
		return
	}


	//查询的总数
	total, _ := supplier.Total(filter)

	resData := models.ResponseSupplierData{}
	resData.Suppliers = suppliers
	resData.Total = total
	resData.Pages = total/size + 1
	resData.Size = size
	resData.CurrentPage = page

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

	supplier.ComID = com.ComId

	//collection := models.Client.Collection("supplier")
	//result := models.Supplier{}
	if !ENABLESAMESUPPLIER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		if supplier.CheckExist() {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该供应商已经存在",
			})
			return
		}
	}
	supplier.ID = int64(getLastID("supplier"))

	err = supplier.Insert()

	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
		return
	}
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

	updateSupplier.ComID = com.ComId

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的supplier_id相等,如果有不相等的，则返回
	// 如果只有相等的supplier_id, 则允许修改

	if !updateSupplier.UpdateCheck() {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Supplier name exist",
		})
		return
	}
	if err := updateSupplier.Update(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Supplier update failed",
		})
		return
	}

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

	supplier := models.Supplier{
		ComID: com.ComId,
		ID: d.ID,
	}
	if err := supplier.Delete(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Customer delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Supplier delete succeeded",
	})
}