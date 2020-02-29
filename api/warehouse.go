package api

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/service"
	"jxc/util"
	"time"

	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)

// 仓库名和仓库地址是否可以重复
const ENABLESAMEWAREHOUSE = false

func AllWarehouses(c *gin.Context) {
	// 根据域名得到com_id
	//用postman调试的时候需要注释
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req models.WarehouseReq
	var warehouses []models.Warehouse

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

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

	if req.ID > 0 {
		filter["warehouse_id"] = bson.M{"$eq": req.ID}
	}

	// Name string `form:"warehouse_name"` //模糊搜索
	if req.Name != "" {
		filter["warehouse_name"] = bson.M{"$regex": req.Name}
	}

	// Address string `form:"warehouse_address"` //模糊搜索
	if req.Address != "" {
		filter["warehouse_address"] = bson.M{"$regex": req.Address}
	}
	// 搜索管理员
	if req.Manager != 0 {
		filter["warehouse_admin_id"] = req.Manager
	}
	// 搜索职员
	if req.Stuff != 0 {
		// 获取职员对应的仓库
		warehouseStuff := models.WarehouseStuff{}
		var warehouses_ids []int64
		stuffsCollection := models.Client.Collection("warehouse_stuffs")
		stuffsCur, _ := stuffsCollection.Find(context.TODO(), bson.M{"com_id": com.ComId, "user_id": req.Stuff})
		for stuffsCur.Next(context.TODO()) {
			_ = stuffsCur.Decode(&warehouseStuff)
			warehouses_ids = append(warehouses_ids, warehouseStuff.WarehouseId)
		}
		warehouses_ids = util.RemoveRepeatedElementInt64(warehouses_ids)
		filter["warehouse_id"] = bson.M{"$in": warehouses_ids}
	}

	// 每个查询都要带着com_id
	//com_id, _ := strconv.Atoi(com.ComId)
	filter["com_id"] = com.ComId
	// all conditions are set then start searching
	collection := models.Client.Collection(("warehouse"))
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}
	var warehouses_ids []int64
	for cur.Next(context.TODO()) {
		var result models.Warehouse
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding recording: ", err)
			return
		}
		warehouses = append(warehouses, result)
		// 获取这批仓库的id
		warehouses_ids = append(warehouses_ids, result.ID)
	}
	// 获取这批仓库的职员
	warehousesStuffs, err := service.FindWarehouseStuffs(warehouses_ids, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 组装职员数据
	for key, val := range warehouses {
		warehouses[key].WarehouseStuff = warehousesStuffs[val.ID]
	}

	//查询的总数
	var total int64
	total, _ = models.Client.Collection("warehouse").CountDocuments(context.TODO(), filter)

	//var wh_stuffs []models.Stuff
	//collection = models.Client.Collection("users")
	//cur, _ = collection.Find(context.TODO(), bson.D{})
	//for cur.Next(context.TODO()) {
	//	var result models.Stuff
	//	if err := cur.Decode(&result); err != nil {
	//		fmt.Println("error while decoding recording: ", err)
	//		return
	//	}
	//	wh_stuffs = append(wh_stuffs, result)
	//}

	resData := models.ResponseWarehouseData{}
	//resData.Warehouses = wh_stuffs
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

// 创建仓库提交的数据
type ReqWarehouse struct {
	ID      int64  `json:"warehouse_id" form:"warehouse_id"`
	Name    string `json:"warehouse_name" form:"warehouse_name"`
	Address string `json:"warehouse_address" form:"warehouse_address"`
	Manager int64  `json:"wh_manager" form:"wh_manager"`
	Config  string `json:"config" form:"config"`
	//仓库员工
	Stuffs []int64 `json:"stuffs" form:"stuff[]"`
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

	var req ReqWarehouse
	user_id := int64(1)

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// TODO: 去人员表中更改对应的仓库权限

	collection := models.Client.Collection("warehouse")
	if !ENABLESAMEWAREHOUSE { //仓库重名检测，
		var result models.Warehouse
		filter := bson.M{}
		filter["com_id"] = com.ComId

		filter["warehouse_name"] = req.Name
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Name != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该仓库已经存在",
			})
			return
		}
	}
	if !ENABLESAMEWAREHOUSE { //仓库地址重名检测，
		var result models.Warehouse
		filter := bson.M{}
		filter["com_id"] = com.ComId

		filter["warehouse_address"] = req.Address
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Name != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该仓库地址已经存在",
			})
			return
		}
	}

	// 管理员id不能出现在职员中
	if util.InArrayInt64(req.Stuffs, req.Manager) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "管理员id不能出现在职员数组中！",
		})
		return
	}
	var user_ids []int64
	user_ids = append(user_ids, req.Manager)
	// 取出所有用户id
	for _, val := range req.Stuffs {
		user_ids = append(user_ids, val)
	}

	// 获取用户信息
	userInfo, err := service.FindUser(user_ids, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 组装仓库数据
	_, ok := userInfo[req.Manager]
	if !ok {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "查找不到该仓库管理员！",
		})
		return
	}
	warehouse_id, _ := util.GetTableId("warehouse")
	warehouse := models.Warehouse{
		ID:                 warehouse_id,
		ComID:              com.ComId,
		Name:               req.Name,
		Address:            req.Address,
		WarehouseAdminId:   req.Manager,
		WarehouseAdminName: userInfo[req.Manager].Username,
		Phone:              userInfo[req.Manager].Phone,
		Config:             req.Config,
		CreateAt:           time.Now().Unix(),
		CreateBy:           user_id,
		ModifyAt:           0,
		ModifyBy:           0,
	}

	// 组装仓库员工数据
	var stuffs []interface{}
	for _, val := range req.Stuffs {
		_, ok = userInfo[val]
		if !ok {
			if !ok {
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"user_id": val},
					Msg:  "查找不到这名用户！",
				})
				return
			}
			return
		}
		stuffs = append(stuffs, models.WarehouseStuff{
			ComID:         com.ComId,
			UserId:        userInfo[val].UserID,
			Username:      userInfo[val].Username,
			WarehouseId:   warehouse_id,
			WarehouseName: warehouse.Name,
		})
	}

	err = service.AddWarehouse(warehouse, stuffs)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	data, _ := service.FindOneWarehouse(warehouse_id, com.ComId)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: data,
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

	var req ReqWarehouse
	user_id := int64(1)

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 仓库名，仓库地址是否可以重复
	//com_id, _ := strconv.Atoi(com.ComId)
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["warehouse_name"] = req.Name
	collection := models.Client.Collection("warehouse")
	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempWarehouse models.Warehouse
		err := cur.Decode(&tempWarehouse)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "无法修改该仓库",
			})
			return
		}
		if tempWarehouse.ID != req.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的仓库名已经存在",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["warehouse_address"] = req.Address

	cur, err = collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempWarehouse models.Warehouse
		err := cur.Decode(&tempWarehouse)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "无法修改该仓库",
			})
			return
		}
		if tempWarehouse.ID != req.ID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的仓库地址已经存在",
			})
			return
		}
	}

	// 管理员id不能出现在职员中
	if util.InArrayInt64(req.Stuffs, req.Manager) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "管理员id不能出现在职员数组中！",
		})
		return
	}
	var user_ids []int64
	user_ids = append(user_ids, req.Manager)
	// 取出所有用户id
	for _, val := range req.Stuffs {
		user_ids = append(user_ids, val)
	}

	// 获取用户信息
	userInfo, err := service.FindUser(user_ids, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 组装仓库数据
	_, ok := userInfo[req.Manager]
	if !ok {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "查找不到该仓库管理员！",
		})
		return
	}

	warehouse := bson.M{
		"warehouse_name":       req.Name,
		"warehouse_address":    req.Address,
		"warehouse_admin_id":   req.Manager,
		"warehouse_admin_name": userInfo[req.Manager].Username,
		"phone":                userInfo[req.Manager].Phone,
		"config":               req.Config,
		"modify_at":            time.Now().Unix(),
		"modify_by":            user_id,
	}

	// 组装仓库员工数据
	var stuffs []interface{}
	for _, val := range req.Stuffs {
		_, ok = userInfo[val]
		if !ok {
			if !ok {
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"user_id": val},
					Msg:  "查找不到这名用户！",
				})
				return
			}
			return
		}
		stuffs = append(stuffs, models.WarehouseStuff{
			ComID:         com.ComId,
			UserId:        userInfo[val].UserID,
			Username:      userInfo[val].Username,
			WarehouseId:   req.ID,
			WarehouseName: req.Name,
		})
	}

	err = service.UpdateWarehouse(req.ID, com.ComId, warehouse, stuffs)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新失败",
		})
		return
	}
	data, _ := service.FindOneWarehouse(req.ID, com.ComId)
	//fmt.Println("Update result: ", result)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: data,
		Msg:  "Warehouse update succeeded",
	})

}

type WarehouseService struct {
	ID int64 `json:"warehouse_id" form:"warehouse_id"`
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

	var d WarehouseService

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
		Msg:  "Warehouse delete succeeded",
	})
}

// 获取仓库详情，有哪些商品，多少库存
func WarehouseDetail(c *gin.Context) {
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

	type ReqWarehouseDetail struct {
		models.BaseReq
		WarehouseId int64  `json:"warehouse_id" form:"warehouse_id"` // 仓库id
		Type        string `json:"type" form:"type"`                 // 搜索类型
		ProductId   int64  `json:"product_id" form:"product_id"`     // 商品id
	}

	var req ReqWarehouseDetail

	// 获取请求数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 获取仓库信息
	warehouse, err := service.FindOneWarehouse(req.WarehouseId, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 组装搜索条件


	// 根据product字段，去获取库存信息
	wosProduct, err := service.GetProductWos(warehouse.Product, com.ComId, req.WarehouseId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: wosProduct,
		Msg:  "",
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
