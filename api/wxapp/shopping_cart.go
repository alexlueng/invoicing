package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/api"
	"jxc/models"
	"jxc/models/wxapp"
	"jxc/serializer"
	"net/http"
	"strings"
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
	ItemID      int64  `json:"item_id"`
	OpenID      string `json:"open_id"`
}

// 如果用户没有购物车 则新建一个
// 用户注册之后就为它创建一个
func CreateCart(com_id, user_id int64, open_id string) (cart wxapp.UserCart, err error) {
	collection := models.Client.Collection("user_cart")
	cart = wxapp.UserCart{
		ComID:      com_id,
		CustomerID: user_id,
		CartID:     api.GetLastID("user"),
		OpenID:     open_id,
	}
	_, err = collection.InsertOne(context.TODO(), cart)
	return
}

// 检查用户购物车是否存在
func CheckMiniappUserCart(com_id, customer_id int64, open_id string) bool {

	collection := models.Client.Collection("user_cart")
	filter := bson.M{}
	filter["com_id"] = com_id
	filter["customer_id"] = customer_id
	filter["open_id"] = open_id
	var userCart wxapp.UserCart
	err := collection.FindOne(context.TODO(), filter).Decode(&userCart)
	if err != nil { // 用户没有cart
		return false
	}

	return true
}

func AddToCart(c *gin.Context) {

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

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 购物车表中查看是否已经有这个商品
	// 如果没有，则新建一条记录，如果有，则将商品数量加1
	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["product_id"] = cartSrv.ProductID
	filter["cart_id"] = cartSrv.CartID
	filter["open_id"] = cartSrv.OpenID // TODO：从token中解出来
	filter["is_delete"] = false
	var cartItem wxapp.CartItem
	err = collection.FindOne(context.TODO(), filter).Decode(&cartItem)
	if err != nil { // 用户新增的购物车项
		cartItem.CartID = cartSrv.CartID
		cartItem.ProductID = cartSrv.ProductID
		cartItem.ComID = com.ComId
		cartItem.IsDelete = false
		cartItem.CreateAt = time.Now().Unix()
		cartItem.Num = 1
		cartItem.ProductName = cartSrv.ProductName
		cartItem.IsAvaliable = true
		cartItem.ItemID = api.GetLastID("cart_item")
		cartItem.OpenID = cartSrv.OpenID

		// 找到这个商品的价格
		var product models.Product
		collection = models.Client.Collection("product")
		filter = bson.M{}
		filter["com_id"] = com.ComId
		filter["product_id"] = cartSrv.ProductID

		err := collection.FindOne(context.TODO(), filter).Decode(&product)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Find image error",
			})
			return
		}

		cartItem.Price = product.DefaultPrice
		// 找到这个商品的图片
		var images []models.Image
		filter = bson.M{}
		filter["com_id"] = com.ComId
		filter["product_id"] = cartSrv.ProductID
		collection = models.Client.Collection("image")
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Find image error",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var res models.Image
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode image error",
				})
				return
			}
			images = append(images, res)
		}
		if len(images) > 0 {
			cartItem.Thumbnail = images[0].CloudPath
		}

		collection = models.Client.Collection("cart_item")
		_, err = collection.InsertOne(context.TODO(), cartItem)
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

	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$inc": bson.M{"num": 1}})
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

// 购物车移除商品
func RemoveFromCart(c *gin.Context) {

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

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 找到这条记录，将这个商品的数据减1
	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["cart_id"] = cartSrv.CartID
	filter["item_id"] = cartSrv.ItemID
	filter["product_id"] = cartSrv.ProductID
	var item wxapp.CartItem
	if err := collection.FindOne(context.TODO(), filter).Decode(&item); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "decode cart item error",
		})
		return
	}

	if item.Num > 0 {
		_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$inc": bson.M{"num": -1}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "update cart item error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "update cart item",
	})
}

// 购物车商品
func DeleteCartItem(c *gin.Context) {

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

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 找到这条记录，将这个商品的数据减1
	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["cart_id"] = cartSrv.CartID
	filter["item_id"] = cartSrv.ItemID
	filter["product_id"] = cartSrv.ProductID

	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"is_delete": true}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "update cart item error",
		})
		return
	}


	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "update cart item",
	})
}


// 清空购物车，将这个购物车中的商品is_avaliable字段置为false
func ClearCart(c *gin.Context) {

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

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["cart_id"] = cartSrv.CartID
	//filter["item_id"] = cartSrv.ItemID
	filter["open_id"] = cartSrv.OpenID
	_, err = collection.UpdateMany(context.TODO(), filter, bson.M{"$set": bson.M{"is_avaliable": false}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "clear user cart error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "clear user cart success",
	})
}

func ListCart(c *gin.Context) {
	//token := c.GetHeader("Access-Token")
	//claims, _ := auth.ParseToken(token)
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

	var cartSrv CartService
	if err := c.ShouldBindJSON(&cartSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("cart_item")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["cart_id"] = cartSrv.CartID
	//filter["item_id"] = cartSrv.ItemID
	filter["open_id"] = cartSrv.OpenID
	filter["is_avaliable"] = true
	filter["is_delete"] = false
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "can't find user cart item",
		})
		return
	}
	var items []wxapp.CartItem
	for cur.Next(context.TODO()) {
		var res wxapp.CartItem
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode user cart item err",
			})
			return
		}
		items = append(items, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "user cart item",
		Data: items,
	})
	return
}
