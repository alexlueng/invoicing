package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"net/http"
	"time"
)

// 处理订单实例的各个函数

//1、未发货
//2、已收货
//3、已发货
//4、部分收货
//5，审核通过
//6，审核不通过

// 销售子订单实例列表
func AllCustomerSubOrderInstance(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	type reqData struct {
		SubOrderSn string `json:"sub_order_sn" form:"sub_order_sn"` // 子订单id
		SubOrderId int64  `json:"sub_order_id" form:"sub_order_id"` //子订单号
	}

	var req reqData
	// 接收数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	instanceArr, err := service.FindSubOrderInstance(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: instanceArr,
		Msg:  "",
	})

}

// 销售子订单实例发货
func CustomerSubOrderInstanceShipped(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	type reqData struct {
		SubOrderSn   string `json:"sub_order_sn" form:"sub_order_sn"`   // 子订单id
		SubOrderId   int64  `json:"sub_order_id" form:"sub_order_id"`   //子订单号
		InstanceId   int64  `json:"instance_id" form:"instance_id"`     //实例id
		DeliveryCom  int64  `json:"delivery_com" form:"delivery_com"`   //配送方式id
		DeliveryCode string `json:"delivery_code" form:"delivery_code"` // 快递单号
	}
	var req reqData
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}
	// 获取实例信息
	instance, err := service.FindOneInstance(req.InstanceId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 获取配送方式
	delivery, err := service.FindOneDelivery(req.DeliveryCom, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["instance_id"] = instance.InstanceId
	// 修改配送方式
	// 订单状态
	// place_type
	instance.DeliveryCom = delivery.DeliveryId // 记录配送方式id，方便查找配送配置
	instance.Delivery = delivery.DeliveryCom   //配送方式

	instance.DeliveryCode = req.DeliveryCode
	instance.ShipTime = time.Now().Unix()
	instance.Status = 2 // 订单实例的状态
	instance.PlaceType = 2
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": instance})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 修改总订单状态为已发货
	collection = models.Client.Collection("customer_order")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_sn"] = instance.SrcOrderSn
	_, err = collection.UpdateOne(context.TODO(),filter, bson.M{"$set" : bson.M{"status" : 2}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Data: instance,
		Msg:  "发货成功！",
	})
}

// 销售子订单实例确认收货
func CustomerSubOrderInstanceConfirm(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 提交参数
	type reqData struct {
		InstanceId int64 `json:"instance_id" form:"instance_id"` //实例id
	}
	var req reqData
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}
	// 获取实例信息
	instance, err := service.FindOneInstance(req.InstanceId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 修改订单状态
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["instance_id"] = instance.InstanceId
	// 修改配送方式
	// 订单状态
	// place_type
	instance.ConfirmTime = time.Now().Unix()
	instance.Status = 3
	instance.PlaceType = 3
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": instance})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	var product models.Product
	collection = models.Client.Collection("product")
	err = collection.FindOne(context.TODO(), bson.D{{"product_id", instance.ProductID}}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	err = service.UpdateProductStock(instance.ProductID, product.Stock+instance.Amount, claims.ComId)
	if err != nil {
		fmt.Println("Update product stock err: ", err)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: instance,
		Msg:  "发货成功！",
	})
}

// 销售子订单实例列表审核
// 审核不通过的时候需要填写不通过的理由
// 这里只需前端多提两个参数,一个是订单审核是否通过，一个是不通过的时候填写的理由

func CustomerSubOrderInstanceCheck(c *gin.Context) {

	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 提交参数, 需要传是否审核通过
	type reqData struct {
		InstanceId int64  `json:"instance_id" form:"instance_id"` //实例id
		Pass       bool   `json:"pass" form:"pass"`               //是否通过
		Reason     string `json:"reason" form:"reason"`           // 审核不通过的理由
	} //
	var req reqData
	// 接收数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	// 获取实例信息
	instance, err := service.FindOneInstance(req.InstanceId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 修改订单状态
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["instance_id"] = instance.InstanceId
	instance.CheckTime = time.Now().Unix()

	if req.Pass {
		// 审核通过
		instance.Status = 4
	} else {
		instance.Status = 5
		instance.FailedReason = req.Reason
	}

	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": instance})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: instance,
		Msg:  "审核完成！",
	})
}

// 采购子订单确认收货
func SupplierSubOrderInstanceConfirm(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	type reqData struct {
		SubOrderSn string `json:"order_sub_sn"` //子订单号
		SubOrderId int64  `json:"order_sub_id"` // 子订单id
		WarehouseID int64 `json:"warehouse_id"`
	}
	var req reqData
	// 接收数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}
	// 根据采购子订单id获取实例
	instance, err := service.SubOrderIdFindOneInstance(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "未能查找到该订单信息",
		})
		return
	}
	// 获取采购子订单信息
	subOrder, err := service.FindSupplierSubOrderByID(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "未能查找到该订单信息",
		})
		return
	}

	// 修改实例状态
	collection := models.Client.Collection("goods_instance")
	instance.ConfirmTime = time.Now().Unix()
	instance.Status = 4
	// 如果是仓库，则修改 placeType
	if instance.DestType == 3 {
		instance.PlaceType = 5
	}
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["instance_id"] = instance.InstanceId
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": instance})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 修改子订单状态
	subOrder.State = 2
	subOrder.ConfirmBy = time.Now().Unix()
	filter1 := bson.M{}
	filter1["com_id"] = claims.ComId
	filter1["order_sub_id"] = subOrder.SubOrderId
	_, err = models.Client.Collection("supplier_sub_order").UpdateOne(context.TODO(), filter1, bson.M{"$set": subOrder})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 修改总订单状态
	collection = models.Client.Collection("supplier_order")
	ordFilter := bson.M{}
	ordFilter["com_id"] = claims.ComId
	ordFilter["order_id"] = subOrder.OrderId
	_, err = collection.UpdateOne(context.TODO(), ordFilter, bson.M{"$set" : bson.M{"status" : 2}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 修改仓库库存情况
	collection = models.Client.Collection("warehouse_product")
	wFilter := bson.M{}
	wFilter["com_id"] = claims.ComId
	wFilter["product_id"] = instance.ProductID
	wFilter["warehouse_id"] = req.WarehouseID
	_, err = collection.UpdateOne(context.TODO(), wFilter, bson.M{
		"$inc" : bson.M{"un_stock" : -instance.Amount, "current_stock" : instance.Amount}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: subOrder,
		Msg:  "确认收货成功！",
	})
}

// 采购子订单审核通过，并增加商品库存
func SupplierSubOrderInstanceCheck(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 一个采购子订单对应一条实例，所以提交采购子订单号可以获取对应实例
	// 提交参数
	type reqData struct {
		SubOrderSn string `json:"order_sub_sn" form:"order_sub_sn"` //子订单号
		SubOrderId int64  `json:"order_sub_id" form:"order_sub_id"` // 子订单id
		Pass       bool   `json:"pass" form:"pass"`                 //是否通过
		Reason     string `json:"reason" form:"reason"`             // 审核不通过的理由
	}

	var req reqData
	// 接收数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 获取采购子订单信息
	subOrder, err := service.FindSupplierSubOrderByID(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "未能查找到该订单信息",
		})
		return
	}

	// 修改子订单状态
	if req.Pass {
		subOrder.State = 5 // 审核通过
	} else {
		subOrder.State = 6 // 审核不通过
		subOrder.FailReason = req.Reason
	}

	subOrder.CheckAt = time.Now().Unix()
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_sub_id"] = subOrder.SubOrderId
	_, err = models.Client.Collection("supplier_sub_order").UpdateOne(context.TODO(), filter, bson.M{"$set": subOrder})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	// 增加商品库存
	collection := models.Client.Collection("product")
	_, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", subOrder.ProductID}}, bson.M{
		"$inc": bson.M{"stock": subOrder.ProductNum}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: subOrder,
		Msg:  "审核完成！",
	})
}
