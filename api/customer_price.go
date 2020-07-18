package api

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"time"

	"github.com/gin-gonic/gin"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

// AddCustomerPrice 操作的是customer_product_price这张表
// 主要有两个地方使用：1.售价管理页面 2.客户下订单时没有对应的售价
func AddCustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerProductPrice models.CustomerProductPrice
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &customerProductPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	if customerProductPrice.DefaultPrice > 0 {
		// 修改默认价格
		// 在商品表，客户商品价格表中都要修改
		// 在客户商品价格表中，把原来的默认价格记为false,再插入一条新的记录

		oldRecord ,err := models.SelectCustomerProductPriceByComIDAndProductID(claims.ComId, customerProductPrice.ProductID, true)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find customer price",
			})
			return
		}

		// 将旧记录设为false
		_, err = models.UpdateCustomerProductPriceValidStatus(claims.ComId, customerProductPrice.ProductID, true, false)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't update old record",
			})
			return
		}

		// 加入一条新记录
		newRecord := oldRecord
		newRecord.Price = customerProductPrice.Price
		err = newRecord.Add()
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert new customer product record",
			})
			return
		}

		err = models.UpdateProductDefaultPriceByProductID(customerProductPrice.ProductID, customerProductPrice.DefaultPrice)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't update product",
			})
			return
		}

		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "update default price success",
		})
		return
	}

	customerProductPrice.ComID = claims.ComId
	// 加上一个时间戳，以及一个有效值
	timestamp := time.Now().Unix()
	customerProductPrice.CreateAt = timestamp
	customerProductPrice.IsValid = true

	// 找到此商品上一个有效价格记录，如果有，则把它设置为无效
	// 找到这个商品的默认价格
	res, err := models.GetProductByID(claims.ComId, customerProductPrice.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}
	customerProductPrice.DefaultPrice = res.DefaultPrice

	_, err = models.SelectCustomerProductPriceByComIDAndCustomerIDAndProductID(claims.ComId, customerProductPrice.CustomerID, customerProductPrice.ProductID, true)
	if err != nil {
		//没有找到这个记录，说明这个客户价格是新增的
		//保存这条记录，更新product表中的cus_product字段

		err = customerProductPrice.Add()
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "添加记录错误",
			})
			return
		}

		// 更新商品客户列表，把客户id追加到cus_price数组中
		err = models.UpdateCusPriceByProductID(customerProductPrice.ProductID, customerProductPrice.CustomerID)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert customer product pricec",
			})
			return
		}

	} else {
		// 找到了旧记录
		// 把旧记录的is_valid字段更新为false,然后插入这条记录
		_, err := models.UpdateCustomerProductPriceValidStatusWithCustomerID(claims.ComId, customerProductPrice.ProductID, customerProductPrice.CustomerID, true, false)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "添加记录错误",
			})
			return
		}
		err = customerProductPrice.Add()
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "添加记录错误",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Insert record succeeded",
	})
	return
}

func ListCustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.CustomerProductPriceReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)
	allProducts, err := models.SelectProductListByComID(claims.ComId, req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get products",
		})
		return
	}

	responseData := make(map[string]map[string]interface{})

	// 根据商品id得到客户名和售价
	// 在商品表中维护一个售价客户id,刚可省去一次循环查找数据库的工作
	// 可以直接从商品表中的cus_price字段中得到已有售价记录的客户id
	allProductsID := []int64{}
	for _, product := range allProducts {
		allProductsID = append(allProductsID, product.ID)
		responseData[product.Product] = make(map[string]interface{})
	}

	filter := bson.M{}
	filter["com_id"] = claims.ComId
	// 按商品名字去搜索
	// TODO: 可以优化这个流程，因为这里只选择一种商品，所以不用循环整个product表了
	if req.Product != "" {
		filter["product_name"] = bson.M{"$regex": req.Product}
	}
	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$regex": req.CustomerName,}
	}
	if len(allProductsID) > 0 {
		filter["product_id"] = bson.M{"$in": allProductsID}
	}
	filter["is_valid"] = true

	csOrdPriceList, err := models.SelectMultiplyCustomerProductPriceByConditoin(filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find customer product price",
		})
		return
	}

	for _, res := range csOrdPriceList.CustomerProductPriceList {
		responseData[res.Product]["product_id"] = res.ProductID
		if responseData[res.Product]["customer_price"] == nil {
			responseData[res.Product]["customer_price"] = []models.CustomerProductPrice{} //make(map[string]models.CustomerProductPrice)
		}
		if res.CustomerID == 0 {
			if responseData[res.Product]["default_price"] == nil {
				responseData[res.Product]["default_price"] = models.CustomerProductPrice{}
			}
			responseData[res.Product]["default_price"] = res
			continue
		}
		responseData[res.Product]["customer_price"] = append(responseData[res.Product]["customer_price"].([]models.CustomerProductPrice), res)
	}

	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$eq": "default"}
		curtomerProductPrice, err := models.SelectMultiplyCustomerProductPriceByConditoin(filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find customer product price",
			})
			return
		}
		for _, csPrice := range curtomerProductPrice.CustomerProductPriceList {
			if responseData[csPrice.Product]["default_price"] == nil {
				responseData[csPrice.Product]["default_price"] = models.CustomerProductPrice{} //make(map[string]models.CustomerProductPrice)
			}
			responseData[csPrice.Product]["default_price"] = csPrice
		}

	}

	var total int
	// TODO: find a good way to calculator total
	total = len(responseData)

	res := models.ResponseCustomerProductPriceData{}
	res.PriceTable = responseData
	res.Size = int(req.Size)
	res.Pages = int(req.Page)
	res.CurrentPage = int(total)/int(req.Size) + 1
	res.Total = int(total)

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get all products",
		Data: res,
	})

}

func DeleteCustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.CustomerProductPriceReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &req)

	_, err := models.UpdateCustomerProductPriceValidStatusWithCustomerID(claims.ComId, req.ProductID, req.CustomerID, true, false)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update customer product price",
		})
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete customer price success",
	})

}
