package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"jxc/util"
	"net/http"
)

// 商品库存实例相关接口

// 商品库存提交的数据
type ReqProductStock struct {
	Product int64 `json:"product"`
}

// 创建库存实例提交的数据
type ReqCreateStock struct {
	Type             int64  `json:"type"`               // 记录类型
	Product          int64  `json:"product"`            // 商品id
	ProductUnitPrice string `json:"product_unit_price"` // 商品单价
	Num              int64  `json:"num"`                // 商品数量
	WarehouseId      int64  `json:"warehouse_id"`       // 仓库id
	OrderSn          string `json:"order_sn"`           // 订单号

}

// 获取商品在仓库中的库存
// 返回有这个商品库存的仓库
func GetProductStock(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	req := ReqProductStock{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	productCount, err := service.GetProductInfoOfWarehouse(req.Product, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	warehouseProductInfo, err := models.ProductInfoOfWarehouse(req.Product, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	warehouses, err := service.FindWarehouse(claims.ComId)

	fmt.Println(warehouses)

	fmt.Println(warehouseProductInfo)

	// 拼接返回数据
	c.JSON(http.StatusOK, serializer.Response{
		Code:  serializer.CodeSuccess,
		Data:  productCount,
	})

}

// 创建库存实例，凭空多出的库存
func CreateWosInstance(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// TODO 用户id设置为1
	req := service.WarehouseData{}
	var insertData []interface{}
	// 验证提交过来的数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	// 获取仓库信息
	warehouse, err := service.FindOneWarehouse(req.WarehouseId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 获取商品信息
	product, err := service.FindOneProduct(req.Product, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	instanceId, _ := util.GetTableId("instance")
	instance := models.GoodsInstance{
		InstanceId:        instanceId,
		ComID:             claims.ComId,
		DestType:          3, // 去向仓库
		DestId:            warehouse.ID,
		DestTitle:         warehouse.Name,
		PlaceId:           warehouse.ID,
		ProductID:         req.Product,
		Product:           product.Product,
		CustomerPrice:     0,
		Amount:            req.Num,
		Status:            4, //已确认
	}
	insertData = append(insertData, instance)

	// 整理仓库中商品种类数据
	warehouse.Product = append(warehouse.Product, product.ProductID)
	// 去重
	warehouse.Product = util.RemoveRepeatedElementInt64(warehouse.Product)

	err = service.AddGoodsInstance(insertData)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 修改查看商品种类
	err = service.UpdateWarehouseProduct(warehouse.ID, claims.ComId, warehouse.Product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	// 修改商品表显示库存
	err = service.UpdateProductStock(product.ProductID, (product.Stock + req.Num), claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
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

type ReqAllWosInstance struct {
	models.BaseReq
	WarehouseId int64  `json:"warehouse_id" form:"warehouse_id"` // 仓库id
	Type        string `json:"type" form:"type"`                 // 搜索类型
	ProductId   int64  `json:"product_id" form:"product_id"`     // 商品id
}

// 仓库实例列表
func AllWosInstance(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ReqAllWosInstance
	var instance models.GoodsInstance
	var instances []models.GoodsInstance

	// 获取请求数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	filter := bson.M{}
	filter["com_id"] = claims.ComId
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
