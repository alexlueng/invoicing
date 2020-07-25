package wxapp

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/conf"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gitee.com/xiaochengtech/wechat/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/api"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
)

// 用户提交订单并处理
// 以及用户订单相关接口

// 提交订单
// 查看订单详情
// 订单各种状态的筛选

// 页面点击去结算 -> 选择收货地址，没有则让用户新建一个 -> 提交购物车商品 -> 生成待支付订单，等待用户支付 -> 收到支付结果后 -> 通知仓库准备发货
//              -> 也可以选择自提                                                               -> 若用户未付款，刚订单会在10分钟内取消

/*
	用户订单状态：
		待支付 1
		待送货 2
		待收货 3
		待评价 4
		退换货 5
		已超时 6
*/

// 订单商品项
type OrderItem struct {
	ItemID      int64   `json:"item_id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Num         int64   `json:"num"`
	Thumbnail   string  `json:"thumbnail"`
}

// 用户提交订单数据
type UserSubmitOrderService struct {
	CustomerID   int64       `json:"customer_id"`            // 客户ID
	CartID       int64       `json:"cart_id" bson:"cart_id"` // 购物车ID
	OpenID       string      `json:"open_id" bson:"open_id"`
	AddressID    int64       `json:"address_id"`    // 地址
	DeliveryID   int64       `json:"delivery_id"`   // 配送方式 物流公司 自提
	Items        []OrderItem `json:"items"`         // 商品项
	DeliveryFee  float64     `json:"delivery_fee"`  // 运费
	CustomerName string      `json:"customer_name"` // 客户名
	PayWay       int64       `json:"pay_way"`       // 支付方式
	Comment      string      `json:"comment"`       // 备注
	TotalPrice   float64     `json:"total_price"`   // 订单总价
}

type CustomerOrderService struct {
	UserID     int64  `json:"user_id"`
	CustomerID int64  `json:"customer_id"`
	OrderID    int64  `json:"order_id"`
	OrderSN    string `json:"order_sn"`
	Status     int64  `json:"status"`
}

// 返回给前端调起微信支付的数据
type PrepayOrderResp struct {
	AppId     string `json:"app_id"`
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
	TotalFee  int    `json:"total_fee"`
	OrderNo   string `json:"order_no"`
}

func getWechatClient() *api.WechatClient {
	config := api.WechatConfig{
		AppId:    os.Getenv("APPID"),
		SubAppId: "",
		MchId:    os.Getenv("MCHID"),
		SubMchId: "",
	}

	apiKey := os.Getenv("APIKEY") // 微信支付上设置的API Key

	wechatClient := api.NewWechatClient(true, api.ServiceTypeNormalDomestic, apiKey, "", config)
	return wechatClient
}

// 返回用户的收货地址
func UserAddresses(c *gin.Context) {

	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	var srv CustomerOrderService
	if err := c.ShouldBindJSON(&srv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("address")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["customer_id"] = srv.UserID
	var addresses []models.Address
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Address
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
		addresses = append(addresses, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "User addresses",
		Data: addresses,
	})

}

/*
	1、商户系统调用统一下单api，从微信后台得到预付编号（prepay_id）
	2、微信小程序从商户系统取得prepay_id等参数，调用 wx.requestPayment()，转到用户付款操作
	3、微信后台向商户系统发送支付结果通知

	小程序支付流程：
	https://pay.weixin.qq.com/wiki/doc/api/wxa/wxa_api.php?chapter=7_4&index=3
*/

// 返回给小程序端的下单参数
type MiniappPayResp struct {
	NonceStr string `json:"nonce_str"`
	PrepayID string `json:"prepay_id"`
	SignType string `json:"sign_type"`
	PaySign  string `json:"pay_sign"`
}

// 提交订单，生成预付单 pre_order
// TODO: 向微信服务器生成预支付订单
func SummitOrder(c *gin.Context) {

	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	var userSubmitOrderSrv UserSubmitOrderService
	if err := c.ShouldBindJSON(&userSubmitOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	preOrder := models.PreOrder{
		ComID:        com.ComId,
		CustomerID:   userSubmitOrderSrv.CustomerID,
		DeliveryID:   userSubmitOrderSrv.DeliveryID,
		CustomerName: userSubmitOrderSrv.CustomerName,
		AddressID:    1, // 测试 暂时写死
		OrderID:      api.GetLastID("customer_order"),
		OrderSN:      GetOrderSN(com.ComId, userSubmitOrderSrv.CustomerID),
		DeliveryFee:  userSubmitOrderSrv.DeliveryFee,
		TotalPrice:   1, // 测试 暂时写死
		PayWay:       userSubmitOrderSrv.PayWay,
		Comment:      userSubmitOrderSrv.Comment,
		IsPay:        false,
		IsCancel:     false,
		IsDelete:     false,
		CreateAt:     time.Now().Unix(),
		ExpireAt:     time.Now().Unix() + 60*10,
		Status:       1, // 未支付
	}

	// 删除购物车中已经下单的商品
	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["cart_id"] = userSubmitOrderSrv.CartID
	filter["open_id"] = userSubmitOrderSrv.OpenID
	for _, item := range userSubmitOrderSrv.Items {
		product := models.CustomerOrderProductsInfo{
			ProductID: item.ProductID,
			Product:   item.ProductName,
			Price:     item.Price,
			Quantity:  item.Num,
			Thumbnail: item.Thumbnail,
		}
		preOrder.Items = append(preOrder.Items, product)
		filter["item_id"] = item.ItemID
		filter["product_id"] = item.ProductID
		_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"is_delete": true}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Params error",
			})
			return
		}
	}

	// 调用微信统一下单接口unifiy_order
	wechatClient := getWechatClient()

	body := api.UnifiedOrderBody{
		NonceStr:       util.RandomString(32),
		Sign:           "",
		Body:           "中民福康-生鲜下单",
		OutTradeNo:     preOrder.OrderSN,
		TotalFee:       1,
		SpbillCreateIP: "119.123.199.102",
		NotifyUrl:      os.Getenv("NOTIFYURL"),
		TradeType:      api.TradeTypeJsApi,
		OpenId:         userSubmitOrderSrv.OpenID,
	}

	resp, err := wechatClient.UnifiedOrder(body)
	if err != nil {
		fmt.Println("Unify order error: ", err.Error())
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "UnifiedOrder request error",
		})
		return
	}

	if resp.ResultCode == "SUCCESS" && resp.ReturnCode == "SUCCESS" { // 预付单请求成功
		//TODO：将这个订单存到redis中，设置过期时间
		collection := models.Client.Collection("pre_order")
		_, err = collection.InsertOne(context.TODO(), preOrder)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert pre_order",
			})
			return
		}

		order, subOrders, err := models.PreOrderToCustomerOrderAndSubOrder(preOrder)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't make orders",
			})
			return
		}

		err = order.Insert()
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
		var iSubs []interface{}
		for _, sub := range subOrders {
			o, ok := sub.(models.CustomerSubOrder)
			if ok {

			}
			o.SubOrderId = api.GetLastID("sub_order")
			o.SubOrderSn = conf.IdWorker.GetOrderSN(order.ComID, order.CustomerID)
			iSubs = append(iSubs, o)
		}
		err = models.MultiplyInsertCustomerSubOrder(iSubs)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}

		// 返回给小程序调起微信支付的数据
		appletResp := PrepayOrderResp{
			AppId:     os.Getenv("APPID"),
			Timestamp: strconv.FormatInt(preOrder.CreateAt, 10),
			NonceStr:  body.NonceStr,
			Package:   "prepay_id=" + resp.PrepayId,
			SignType:  "MD5",
			PaySign:   "",
			TotalFee:  body.TotalFee,
			OrderNo:   body.OutTradeNo,
		}

		appletResp.PaySign = api.GetAppletPaySign(
			resp.AppId,
			appletResp.NonceStr,
			appletResp.Package,
			appletResp.SignType,
			appletResp.Timestamp,
			os.Getenv("APIKEY"))

		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "生成预付订单，请尽快支付",
			Data: appletResp,
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeError,
		Msg:  "订单失败，请重试",
	})
}

// 微信支付回调函数
func WxpayCallback(c *gin.Context) {

	//配置请求头
	c.Header("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("failed to read http body: ", err)
		return
	}

	wechatClient := getWechatClient()
	resp, err := wechatClient.NotifyPay(HandlePayNotify, body)

	// 向微信返回
	c.String(http.StatusOK, resp)
}

// 根据回调结果来处理订单
func HandlePayNotify(body api.NotifyPayBody) error {

	if body.ReturnCode == "FAIL" {
		fmt.Printf("ErrorCode %s and ErrorCodeDes %s", body.ErrCode, body.ErrCodeDes)
		payFailErr := errors.New("用户支付失败")
		return payFailErr
	}

	// 用户支付成功，收款后将订单状态改为待送货
	orderSN := body.OutTradeNo
	fmt.Println("Pay orderSN: ", orderSN)
	collection := models.Client.Collection("customer_order")
	filter := bson.M{}
	filter["order_sn"] = orderSN
	var preOrder models.PreOrder
	err := collection.FindOne(context.TODO(), filter).Decode(&preOrder)
	if err != nil {
		payFailErr := errors.New("没有找到这个订单")
		return payFailErr
	}

	_, err = collection.UpdateOne(context.TODO(), bson.D{{"order_sn", orderSN}}, bson.M{"$set": bson.M{
		"pay_at": time.Now().Unix(), "status": 2, "is_pay": true,
	}})
	if err != nil {
		payFailErr := errors.New("更新订单状态失败")
		return payFailErr
	}

	// TODO：通过短信，邮件，或者语音系统提醒用户

	return nil
}

// 需要一个安全生成订单号的方法
func GetOrderSN(com_id, user_id int64) string {
	return conf.IdWorker.GetOrderSN(com_id, user_id)
}

// 订单超时, 前台传回来，也可以从redis超时判断
func OrderExpire(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 更新订单为取消状态
	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = customerOrderSrv.OrderID
	filter["order_sn"] = customerOrderSrv.OrderSN
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"is_cancel": true,
		"status": 2}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't cancel order",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Cancel order",
	})
	return
}

// 收到微信的支付成功回调，将订单从预付状态改变成待送货
func PayPreOrder(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 更新订单为待发货状态
	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = customerOrderSrv.OrderID
	filter["order_sn"] = customerOrderSrv.OrderSN
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"status": 3, "is_pay": true}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update PreOrder",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "order is going to be shipped",
	})
	return
}

// 后台处理订单发货，将订单从待送货变成待收货
func ShipOrder(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 更新订单为用户待收货状态
	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = customerOrderSrv.OrderID
	filter["order_sn"] = customerOrderSrv.OrderSN
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"status": 4, "ship_at": time.Now().Unix()}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't ship PreOrder",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "order is shipped",
	})
	return
}

// 用户删除订单
func DeleteOrder(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 更新订单为用户待收货状态
	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = customerOrderSrv.OrderID
	filter["order_sn"] = customerOrderSrv.OrderSN
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"status": 5, "finish_at": time.Now().Unix()}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't finish PreOrder",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "order is finished",
	})
	return
}

// 用户可以点确定收货，平台也可以根据送达的状态判断订单是否送达
func FinishOrder(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 更新订单为用户待收货状态
	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = customerOrderSrv.OrderID
	filter["order_sn"] = customerOrderSrv.OrderSN
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"status": 5, "finish_at": time.Now().Unix()}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't finish PreOrder",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "order is finished",
	})
	return
}

// 所有订单
func AllOrders(c *gin.Context) {

	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	var customerOrderSrv CustomerOrderService
	if err := c.ShouldBindJSON(&customerOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// TODO: 根据状态来选择订单，加入分页，搜索

	collection := models.Client.Collection("customer_order")
	option := options.Find()
	option.SetSort(bson.D{{"order_time", -1}})
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["customer_id"] = customerOrderSrv.CustomerID
	filter["order_type"] = 1
	if customerOrderSrv.Status > 0 {
		filter["status"] = customerOrderSrv.Status
	}
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get PreOrder",
		})
		return
	}

	var preOrders []models.CustomerOrder
	for cur.Next(context.TODO()) {
		var res models.CustomerOrder
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode PreOrder",
			})
			return
		}
		preOrders = append(preOrders, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "user orders",
		Data: preOrders,
	})
	return
}

// 评价订单
func OrderFeedBack(c *gin.Context) {}

// 申请售后
func ApplyAfterSales(c *gin.Context) {}
