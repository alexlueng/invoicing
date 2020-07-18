package wxapp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"jxc/api"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"

	"gitee.com/xiaochengtech/wechat/util"
	"gitee.com/xiaochengtech/wechat/wxpay"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
		已超时 2
		待送货 3
		待收货 4
		待评价 5
		退换货
*/

// 订单商品项
type OrderItem struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Num         int64   `json:"num"`
}

// 用户提交订单数据
type UserSubmitOrderService struct {
	CustomerID   int64       `json:"customer_id"`   // 客户ID
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

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userSubmitOrderSrv UserSubmitOrderService
	if err := c.ShouldBindJSON(&userSubmitOrderSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	preOrder := models.PreOrder{
		ComID:       claims.ComId,
		CustomerID:  userSubmitOrderSrv.CustomerID,
		DeliveryID:  userSubmitOrderSrv.DeliveryID,
		AddressID:   userSubmitOrderSrv.AddressID,
		OrderID:     api.GetLastID("orders"),
		OrderSN:     GetOrderSN(),
		DeliveryFee: userSubmitOrderSrv.DeliveryFee,
		TotalPrice:  userSubmitOrderSrv.TotalPrice,
		PayWay:      userSubmitOrderSrv.PayWay,
		Comment:     userSubmitOrderSrv.Comment,
		IsPay:       false,
		IsCancel:    false,
		IsDelete:    false,
		CreateAt:    time.Now().Unix(),
		ExpireAt:    time.Now().Unix() + 60*10,
		Status:      1, // 未支付
	}

	for _, item := range userSubmitOrderSrv.Items {
		product := models.CustomerOrderProductsInfo{
			ProductID: item.ProductID,
			Product:   item.ProductName,
			Price:     item.Price,
			Quantity:  item.Num,
		}
		preOrder.Items = append(preOrder.Items, product)
	}

	// 调用微信统一下单接口unifiy_order
	config := wxpay.Config{
		AppId:    "",
		SubAppId: "",
		MchId:    "",
		SubMchId: "",
	}

	apiKey := "xxxxxxxx" // 微信支付上设置的API Key

	client := wxpay.NewClient(false, wxpay.ServiceTypeNormalDomestic, apiKey, "", config)
	fmt.Println(client)

	body := wxpay.UnifiedOrderBody{
		AppId:          "",
		MchId:          "",
		NonceStr:       util.RandomString(32),
		Sign:           "",
		Body:           "",
		OutTradeNo:     preOrder.OrderSN,
		TotalFee:       0,
		SpbillCreateIP: "",
		NotifyUrl:      "",
		TradeType:      wxpay.TradeTypeJsApi,
	}

	//body.Sign = wxpaySign(body, apiKey)

	resp, err := client.UnifiedOrder(body)
	if err != nil {
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

		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "生成预付订单，请尽快支付",
		})
	}

}

// 微信支付计算签名的函数
func wxpaySign(mReq map[string]interface{}, key string) (sign string) {
	fmt.Println("微信支付签名计算, API KEY:", key)
	//STEP 1, 对key进行升序排序.
	sorted_keys := make([]string, 0)
	for k, _ := range mReq {
		sorted_keys = append(sorted_keys, k)
	}

	sort.Strings(sorted_keys)

	//STEP2, 对key=value的键值对用&连接起来，略过空值
	var signStrings string
	for _, k := range sorted_keys {
		value := fmt.Sprintf("%v", mReq[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value + "&"
		}
	}

	//STEP3, 在键值对的最后加上key=API_KEY
	if key != "" {
		signStrings = signStrings + "key=" + key
	}
	//STEP4, 进行MD5签名并且将所有字符转为大写.
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signStrings))
	cipherStr := md5Ctx.Sum(nil)
	upperSign := strings.ToUpper(hex.EncodeToString(cipherStr))

	fmt.Println("Get wxpay sign: ", upperSign)

	return upperSign
}


// 需要一个安全生成订单号的方法
func GetOrderSN() string {
	return ""
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

	// TODO: 根据状态来选择订单，加入分页，搜索

	collection := models.Client.Collection("pre_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["customer_id"] = customerOrderSrv.CustomerID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get PreOrder",
		})
		return
	}

	var preOrders []models.PreOrder
	for cur.Next(context.TODO()) {
		var res models.PreOrder
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
