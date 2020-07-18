package api

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

//允许同名的供应商
const ENABLESAMESUPPLIER = false

func ListSuppliers(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.SupplierReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
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

	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}
	if req.Contacts != "" {
		filter["contacts"] = bson.M{"$regex": req.Contacts}
	}

	if req.Name != "" {
		filter["supplier_name"] = bson.M{"$regex": req.Name}
	}

	filter["com_id"] = claims.ComId

	var suppliers []models.Supplier
	supplier := models.Supplier{}

	suppliers, err = supplier.FindAll(filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find suppliers",
		})
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
		Code: serializer.CodeSuccess,
		Msg:  "Get suppliers",
		Data: resData,
	})
}
func AddSuppliers(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)
	supplier := models.Supplier{}

	_ = json.Unmarshal(data, &supplier)

	supplier.ComID = claims.ComId

	if !ENABLESAMESUPPLIER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		if supplier.CheckExist() {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "该供应商已经存在",
			})
			return
		}
	}
	supplier.ID = GetLastID("supplier")
	supplier.SupplyList = []int64{}

	// 设置供应商初始密码
	supplier.Password = supplier.Phone[5:]

	// 初始化供应商通知方式
	supplier.NotifyWay = append(supplier.NotifyWay, models.Notify{
		Name:  "email",
		Using: false,
	})
	supplier.NotifyWay = append(supplier.NotifyWay, models.Notify{
		Name:  "sms",
		Using: false,
	})
	supplier.NotifyWay = append(supplier.NotifyWay, models.Notify{
		Name:  "wx",
		Using: false,
	})

	err := supplier.Insert()

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't add new supplier",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Supplier create succeeded",
		Data: supplier,
	})
}

func UpdateSuppliers(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	updateSupplier := models.Supplier{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateSupplier)

	updateSupplier.ComID = claims.ComId

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的supplier_id相等,如果有不相等的，则返回
	// 如果只有相等的supplier_id, 则允许修改

	if !updateSupplier.UpdateCheck() {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Supplier name exist",
		})
		return
	}
	if err := updateSupplier.Update(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Supplier update failed",
		})
		return
	}

	collection := models.Client.Collection("supplier_product_price")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["supplier_id"] = updateSupplier.ID
	filter["is_valid"] = true

	_, err := collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"supplier": updateSupplier.SupplierName,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Supplier product price update failed",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Supplier update succeeded",
	})
}

type DeleteSupplierService struct {
	ID int64 `json:"supplier_id"`
}

func DeleteSuppliers(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var d DeleteSupplierService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	supplier := models.Supplier{
		ComID: claims.ComId,
		ID:    d.ID,
	}
	if err := supplier.Delete(); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Customer delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Supplier delete succeeded",
	})
}

