package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"jxc/util"
	"net/http"
	"strings"
)

// 商品库存实例相关接口

// 商品库存提交的数据
type ReqProductStock struct {
	Products []int64 `json:"products" form:"products[]"`
}

// 创建库存实例提交的数据
type ReqCreateWosExamples struct {
	Type             int64  `json:"type"`               // 记录类型
	Product          int64  `json:"product"`            // 商品id
	ProductUnitPrice string `json:"product_unit_price"` // 商品单价
	Num              int64  `json:"num"`                // 商品数量
	WarehouseId      int64  `json:"warehouse_id"`       // 仓库id
	OrderSn          string `json:"order_sn"`           //订单号

}

// 获取商品库存
func ProductWos(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	req := ReqProductStock{}
	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	productCount, err := service.GetProductWos(req.Products, com.ComId, 0)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 拼接返回数据
	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  productCount,
		Msg:   "",
		Error: "",
	})

}

// 创建库存实例，凭空多出的库存
func CreateWosInstance(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	// TODO 用户id设置为1
	//user_id := int64(1)
	req := service.WosExamplesData{}
	var insertData []interface{}
	// 验证提交过来的数据
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

	// 获取商品信息
	product, err := service.FindOneProduct(req.Product, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	instanceId, _ := util.GetTableId("instance")
	instance := models.GoodsInstance{
		InstanceId:        instanceId,
		ComID:             com.ComId,
		Type:              0,
		SrcType:           0, // 没有来源
		SrcId:             0,
		SrcTitle:          "",
		SrcOrderId:        0,// 没有来源，所以来源订单id为0
		SrcOrderSn:        "",
		SrcSubOrderId:     0,
		SrcSubOrderSn:     "",
		DestType:          3, // 去向仓库
		DestId:            warehouse.ID,
		DestTitle:         warehouse.Name,
		DestOrderId:       0,// 发给仓库，所以没有去向订单id
		DestOrderSn:       "",
		DestSubOrderId:    0,
		DestSubOrderSn:    "",
		PlaceType:         7,
		PlaceId:           warehouse.ID,
		SubPlaceId:        0,
		ProductID:         req.Product,
		Product:           product.Product,
		Contacts:          "",
		Receiver:          "",
		ReceiverPhone:     "",
		SupplierPrice:     0,
		CustomerPrice:     0,
		Amount:            0,
		ExtraAmount:       0,
		Delivery:          "",
		DeliveryCode:      "",
		OrderTime:         0,
		ShipTime:          0,
		ConfirmTime:       0,
		PayTime:           0,
		FinishTime:        0,
		Status:            4, //已确认
		SettlementOrderSN: "",
		Settlement:        0,
	}
	insertData = append(insertData, instance)

	err = service.AddGoodsInstance(insertData)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "添加库存记录成功",
	})

}

// 仓库实例日志返回的数据格式
type ResponseWarehouseData struct {
	WosInstances interface{} `json:"wos_instances"`
	Total        int         `json:"total"`
	Pages        int         `json:"pages"`
	Size         int         `json:"size"`
	CurrentPage  int         `json:"current_page"`
}

// 仓库实例列表
func AllWosInstance(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	type ReqAllWosInstance struct {
		models.BaseReq
		WarehouseId int64  `json:"warehouse_id" form:"warehouse_id"` // 仓库id
		Type        string `json:"type" form:"type"`                 // 搜索类型
		ProductId   int64  `json:"product_id" form:"product_id"`     // 商品id
	}

	var req ReqAllWosInstance
	var instance models.GoodsInstance
	var instances []models.GoodsInstance

	// 获取请求数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["place_id"] = req.WarehouseId
	filter["product_id"] = req.ProductId

	// placeType 1 销售-待发货
	// placeType 2 销售-已发货
	// placeType 3 销售-确认收货
	// placeType 4 采购-待收货
	// placeType 5 采购-已收货
	// placeType 6 无端损耗
	// placeType 7 凭空增加
	switch req.Type {
	case "all": // 所有日志
		break
	case "in_wos":
		// 在库中
		// placeType 5 采购-已收货
		filter["place_type"] = 5
		break
	case "not_wosed":
		// placeType 4 采购-待收货
		filter["place_type"] = 4
		break
	case "shipped":
		// placeType 2 销售-已发货
		filter["place_type"] = 2
		break
	case "not_shipped":
		// placeType 1 销售-待发货
		filter["place_type"] = 1
		break
	default:

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
	// 设置分页的默认值
	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)
	req.OrdF = "create_at"
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))

	//1从小到大,-1从大到小
	option.SetSort(bson.D{{req.OrdF, order}})

	collection := models.Client.Collection("goods_instance")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&instance)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}
		instances = append(instances, instance)
	}

	var total int64
	total, err = collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	resData := ResponseWarehouseData{}
	resData.WosInstances = instances
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: resData,
		Msg:  "获取仓库实例日志成功！",
	})
}

// 仓库发货 shipp
func WosShipp(c *gin.Context) {
	// 库存实例修改订单状态- 发货

	// 销售 -> 销售订单
	// 退货 -> 销售订单 暂时没有退货
	// 采购 -> 采购订单

	// 只有销售 - 销售订单仓库才能发货  提交的数据未 采购订单号，商品id

	// 采购 接收到

	// 提交的数据
	// 配送方式
	//

}

// 更新实例状态
