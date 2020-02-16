package api

import (
	"context"
	//"github.com/gin-gonic/gin/internal/json"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
)

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)

const ENABLESAMEWAREHOUSE = false

type Warehouse struct {
	ID int64 `json:"warehouse_id" bson:"warehouse_id"`
	ComID int64 `json:"com_id" bson:"com_id"`
	Name string `json:"warehouse_name" bson:"warehouse_name"`
	Address string `json:"warehouse_address" bson:"warehouse_address"`
	Manager string `json:"wh_manager" bson:"wh_manager"`
}

//用户提交过来的数据
type WarehouseReq struct {
	IdMin int `form:"idmin"` //okid界于[idmin 和 idmax] 之间的数据
	IdMax int `form:"idmax"` //ok
	//本页面的搜索字段 sf固定等于customer_name， key的值为用户提交过来的客户名关键字
	Key  string `form:"key"`              //用户提交过来的模糊搜索关键字
	Sf   string `form:"sf"`               //用户模糊搜索的字段  search field
	Page int64  `json:"page" form:"page"` //ok用户查询的是哪一页的数据
	Size int64  `json:"size" form:"size"` //ok用户希望每页展现多少条数据
	OrdF string `json:"ordf" form:"ordf"` //ok用户排序字段 order field
	Ord  string `json:"ord" form:"ord"`   //ok顺序还是倒序排列  ord=desc 倒序，ord = asc 升序
	TMin int    `form:"tmin"`             //时间最小值[tmin,tmax)
	TMax int    `form:"tmax"`             //时间最大值
	//本页面定制的搜索字段
	Name      string `json:"warehouse_name" form:"warehouse_name"`       //模糊搜索
	Address  string `json:"warehouse_address" form:"warehouse_address"`   //模糊搜索
}

type ResponseWarehouseData struct {
	Warehouses   []Warehouse `json:"warehouses"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}


func AllWarehouses(c *gin.Context) {
	// 根据域名得到com_id
	// 用postman调试的时候需要注释
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req WarehouseReq
	var warehouses []Warehouse

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 设置分页的默认值
	if req.Size < 11 {
		req.Size = 10
	}
	if req.Page < 2 {
		req.Page = 1
	}

	// 设置排序主键
	orderField := []string{"warehouse_id", "com_id", "warehouse_address", "wh_manager", "warehouse_name"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "warehouse_id"
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

	// Reciever string `form:"warehouse_name"` //模糊搜索
	if req.Name != "" {
		filter["warehouse_name"] = bson.M{"$regex": req.Name}
	}

	// Reciever string `form:"warehouse_address"` //模糊搜索
	if req.Address != "" {
		filter["warehouse_address"] = bson.M{"$regex": req.Address}
	}

	// 每个查询都要带着com_id
	//com_id, _ := strconv.Atoi(com.ComId)
	//filter["com_id"] = com_id
	filter["com_id"] = 1
	// all conditions are set then start searching
	collection := models.Client.Collection(("warehouse"))
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result Warehouse
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding recording: ", err)
			return
		}
		warehouses = append(warehouses, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("warehouse").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	resData := ResponseWarehouseData{}
	resData.Warehouses = warehouses
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get warehouses",
		Data: resData,
	})


}

func AddWarehouse(c *gin.Context) {
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

	var warehouse Warehouse
	err = json.Unmarshal(data, &warehouse)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "warehouse create failed",
		})
		return
	}
	collection := models.Client.Collection("warehouse")
	if !ENABLESAMEWAREHOUSE { //仓库重名检测，
		var result Warehouse
		filter := bson.M{}
//		com_id, _ := strconv.Atoi(com.ComId)
		filter["com_id"] = com.ComId

		filter["warehouse_name"] = warehouse.Name
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Name != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该仓库已经存在",
			})
			return
		}
	}

	warehouse.ID = int64(getLastWarehouseID())
	//com_id, _ := strconv.Atoi(com.ComId)
	warehouse.ComID = com.ComId

	insertResult, err := collection.InsertOne(context.TODO(), warehouse)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
		return
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer create succeeded",
	})

}

func UpdateWarehouse(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	updateWarehouse := Warehouse{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &updateWarehouse)
	if err != nil {
		fmt.Println("error while unmarshal the update warehouse data.")
		return
	}

	// 仓库名，仓库地址是否可以重复
	//com_id, _ := strconv.Atoi(com.ComId)
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["warehouse_name"] = updateWarehouse.Name
	collection := models.Client.Collection("warehouse")
	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempWarehouse Warehouse
		err := cur.Decode(&tempWarehouse)
		if err != nil {
			fmt.Println("error found decoding warehouse: ", err)
			return
		}
		if tempWarehouse.ID != updateWarehouse.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的仓库名已经存在",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["warehouse_address"] = updateWarehouse.Address

	cur, err = collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempWarehouse Warehouse
		err := cur.Decode(&tempWarehouse)
		if err != nil {
			fmt.Println("error found decoding warehouse: ", err)
			return
		}
		if tempWarehouse.ID != updateWarehouse.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的仓库地址已经存在",
			})
			return
		}
	}


	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["warehouse_id"] = updateWarehouse.ID
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"warehouse_name": updateWarehouse.Name,
			"wh_manager": updateWarehouse.Manager,
			"warehouse_address": updateWarehouse.Address}})
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
		Msg:  "Warehouse update succeeded",
	})


}

type DeleteWarehouseService struct {
	ID int64 `json:"warehouse_id"`
}

func DeleteWarehouse(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
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
	//com_id, _ := strconv.Atoi(com.ComId)
	filter["com_id"] = com.ComId
	filter["warehouse_id"] = d.ID
	collection := models.Client.Collection("warehouse")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除仓库失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer delete succeeded",
	})
}


type WarehouseCount struct {
	NameField string
	Count     int
}
// 因mongodb不允许自增方法，所以要生成新增客户的id
// 这是极度不安全的代码，因为本程序是分布式的，本程序可能放在多台服务器上同时运行的。
// 需要在交付之前修改正确
func getLastWarehouseID() int {
	var wc WarehouseCount
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", "warehouse"}}).Decode(&wc)
	if err != nil {
		fmt.Println("can't get warehouseID")
		return 0
	}
	collection.UpdateOne(context.TODO(), bson.M{"name": "warehouse"}, bson.M{"$set": bson.M{"count": wc.Count + 1}})
	fmt.Println("customer count: ", wc.Count)
	return wc.Count + 1
}
