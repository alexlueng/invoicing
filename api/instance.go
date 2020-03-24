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
	"strings"
	"time"
)

//

// 销售子订单实例列表
func AllCustomerSubOrderInstance(c *gin.Context) {
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


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
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	// 接收数据
	// 子订单号
	// 实例id
	// 配送方式
	// 快递单号
	type reqData struct {
		SubOrderSn   string `json:"sub_order_sn" form:"sub_order_sn"`   // 子订单id
		SubOrderId   int64  `json:"sub_order_id" form:"sub_order_id"`   //子订单号
		InstanceId   int64  `json:"instance_id" form:"instance_id"`     //实例id
		DeliveryCom  int64  `json:"delivery_com" form:"delivery_com"`   //配送方式id
		DeliveryCode string `json:"delivery_code" form:"delivery_code"` // 快递单号
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
	// 获取实例信息
	instance, err := service.FindOneInstance(req.InstanceId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 获取配送方式
	delivery, err := service.FindOneDelivery(req.DeliveryCom, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
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
	instance.Status = 5
	instance.PlaceType = 2
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
		Msg:  "发货成功！",
	})
}

// 销售子订单实例确认收货
func CustomerSubOrderInstanceConfirm(c *gin.Context) {
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	// 提交参数
	type reqData struct {
		InstanceId int64 `json:"instance_id" form:"instance_id"` //实例id
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
	instance.Status = 2
	instance.PlaceType = 3
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
		Msg:  "发货成功！",
	})
}

// 销售子订单实例列表审核
func CustomerSubOrderInstanceCheck(c *gin.Context) {
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	// 提交参数
	type reqData struct {
		InstanceId int64 `json:"instance_id" form:"instance_id"` //实例id
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
	instance.CheckTime = time.Now().Unix()
	instance.Status = 3
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
		Msg:  "发货成功！",
	})
}

// 采购子订单发货
func SupplierSubOrderShipped(c *gin.Context) {
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	// 一个采购子订单对应一条实例，所以提交采购子订单号可以获取对应实例
	// 提交参数
	type reqData struct {
		SubOrderSn string `json:"sub_order_Sn" form:"sub_order_Sn"` //实例id
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
	// 根据采购子订单号获取实例
	instance, err := service.SubOrderSnFindOneInstance(req.SubOrderSn, claims.ComId)
	if err != nil {
		return
	}
	// 修改订单状态
	collection := models.Client.Collection("goods_instance")
	var filter, update bson.M
	update["ship_time"] = time.Now().Unix()
	update["status"] = 3

	filter["com_id"] = claims.ComId
	filter["instance_id"] = instance.InstanceId
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "发货成功！",
	})
}

// 采购子订单确认收货
func SupplierSubOrderInstanceConfirm(c *gin.Context) {
	// 根据域名获取comid
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	// 一个采购子订单对应一条实例，所以提交采购子订单号可以获取对应实例
	// 提交参数
	type reqData struct {
		SubOrderSn string `json:"order_sub_sn" form:"order_sub_sn"` //子订单号
		SubOrderId int64  `json:"order_sub_id" form:"order_sub_id"` // 子订单id
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
	// 根据采购子订单id获取实例
	instance, err := service.SubOrderIdFindOneInstance(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能查找到该订单信息",
		})
		return
	}
	// 获取采购子订单信息
	subOrder, err := service.SubOrderIdFindOneSupplierSubOrder(req.SubOrderId, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
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
	filter:= bson.M{}
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
	subOrder.State = 4
	subOrder.ConfirmBy = time.Now().Unix()
	filter1 := bson.M{}
	filter1["com_id"] = claims.ComId
	filter1["order_sub_id"] = subOrder.SubOrderId
	_, err = models.Client.Collection("supplier_sub_order").UpdateOne(context.TODO(), filter1, bson.M{"$set": subOrder})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
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

// 采购子订单审核通过
func SupplierSubOrderInstanceCheck(c *gin.Context) {
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

	// 一个采购子订单对应一条实例，所以提交采购子订单号可以获取对应实例
	// 提交参数
	// 一个采购子订单对应一条实例，所以提交采购子订单号可以获取对应实例
	// 提交参数
	type reqData struct {
		SubOrderSn string `json:"order_sub_sn" form:"order_sub_sn"` //子订单号
		SubOrderId int64  `json:"order_sub_id" form:"order_sub_id"` // 子订单id
	}
	var req reqData
	// 接收数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	// 根据采购子订单id获取实例
	instance, err := service.SubOrderIdFindOneInstance(req.SubOrderId, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能查找到该订单信息",
		})
		return
	}
	// 获取采购子订单信息
	subOrder, err := service.SubOrderIdFindOneSupplierSubOrder(req.SubOrderId, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能查找到该订单信息",
		})
		return
	}

	// 修改实例状态
	collection := models.Client.Collection("goods_instance")
	instance.CheckTime = time.Now().Unix()
	instance.Status = 3
	// 如果是仓库，则修改 placeType
	if instance.DestType == 3 {
		instance.PlaceType = 5
	}
	filter := bson.M{}
	filter["com_id"] = com.ComId
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
	subOrder.State = 3
	subOrder.CheckAt = time.Now().Unix()
	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["sub_order_id"] = subOrder.SubOrderId
	_, err = models.Client.Collection("supplier_sub_order").UpdateOne(context.TODO(), filter, bson.M{"$set": subOrder})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: subOrder,
		Msg:  "审核通过！",
	})
}
