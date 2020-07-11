package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/api"
	"jxc/auth"
	"jxc/models"
	"jxc/models/wxapp"
	"jxc/serializer"
	"net/http"
	"time"
)

// 购物车功能接口
// 用户浏览产品加入购物车保存到购物车表中
// 购物车中的商品点击下单进入到订单表中
// 实现思路：接收前台传过来的购物车参数，生成订单
// 大概要接收的参数：商品，数量，总金额

type CartService struct {
	ComID       int64  `json:"com_id"`
	CustomerID  int64  `json:"customer_id"`
	CartID      int64  `json:"cart_id"`
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
}

// 如果用户没有购物车 则新建一个
// 用户注册之后就为它创建一个
func CreateCart(c *gin.Context) {

}

// 检查用户购物车是否存在
func CheckCustomerCart(com_id, customer_id int64) bool {

	collection := models.Client.Collection("user_cart")
	filter := bson.M{}
	filter["com_id"] = com_id
	filter["customer_id"] = customer_id
	var userCart wxapp.UserCart
	err := collection.FindOne(context.TODO(), filter).Decode(&userCart)
	if err != nil { // 用户没有cart
		userCart.CustomerID = customer_id
		userCart.ComID = com_id
		userCart.CartID = api.GetLastID("user_cart")

		_, err := collection.InsertOne(context.TODO(), userCart)
		if err != nil {
			return false
		}
	}

	return true
}

func AddToCart(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 购物车表中查看是否已经有这个商品
	// 如果没有，则新建一条记录，如果有，则将商品数量加一
	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["customer_id"] = cartSrv.ProductID
	filter["cart_id"] = cartSrv.CartID
	var cartItem wxapp.CartItem
	err := collection.FindOne(context.TODO(), filter).Decode(&cartItem)
	if err != nil { // 用户新增的购物车项
		cartItem.CartID = cartSrv.CartID
		cartItem.ProductID = cartSrv.ProductID
		cartItem.ComID = claims.ComId
		cartItem.IsDelete = false
		cartItem.CreateAt = time.Now().Unix()
		cartItem.Num = 1
		cartItem.ProductName = cartSrv.ProductName
		cartItem.IsAvaliable = true
		cartItem.ItemID = api.GetLastID("cart_item")

		_, err := collection.InsertOne(context.TODO(), cartItem)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert new cart item",
			})
			return
		}
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "Create new cart item",
		})
		return
	}

	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$inc" : bson.M{"num" : 1}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update cart item",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update cart item",
	})
}

func RemoveFromCart(c *gin.Context) {

}

func ClearCart(c *gin.Context) {

}
