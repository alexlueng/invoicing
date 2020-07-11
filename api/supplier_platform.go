package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// 供应商操作平台

type SupplierLoginService struct {
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
}

type SupplierService struct {
	SupplierID     int64               `json:"supplier_id"`
	OrderStatus    int64               `json:"order_status"` // 0 全部 1 未处理 2
	SettlementType int64               `json:"settlement_type"`
	DeviceName     string              `json:"device_name"`
	DeviceNumber   string              `json:"device_number"`
	NotifyWay      []models.Notify     `json:"notify_way"`
	OrderID        int64               `json:"order_id"`
	SubOrders      []OrderActualAmount `json:"order_actual_amount"`
}

type OrderActualAmount struct {
	SubOrderID   int64 `json:"sub_order_id"`
	ActualAmount int64  `json:"actual_amount"`
}

type ProductInfoDetail struct {
	ProductID  int64 `json:"product_id"`
	SupplierID int64 `json:"supplier_id"`
}

type SupplierResetPasswordService struct {
	SupplierID  int64  `json:"supplier_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_pass"`
	ConfirmPass string `json:"confirm"`
}

type SupplierProductInfo struct {
	ProudctPrice models.SupplierProductPrice
	Urls         []string
}

// 供应商登录
func SupplierLogin(c *gin.Context) {

	domain := c.Request.Header.Get("Origin")
	fmt.Println("请求域名：", domain[len("http://"):]) // TODO：这里不能这样写，要改成灵活的方式
	com, err := models.GetComIDAndModuleByDomain(domain[len("http://"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	var slSrv SupplierLoginService
	if err := c.ShouldBindJSON(&slSrv); err != nil {
		fmt.Println("param error")
		return
	}

	// 查询数据库里是否有此供应商
	collection := models.Client.Collection("supplier")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["phone"] = slSrv.Telephone

	var supplier models.Supplier
	err = collection.FindOne(context.TODO(), filter).Decode(&supplier)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "您还不是我们的供应商",
		})
		return
	}

	// 比较密码 密码还需要经过加密
	if slSrv.Password != supplier.Password {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "密码错误",
		})
		return
	}

	token, _ := auth.GenerateToken(supplier.SupplierName, supplier.ID, com.ComId, false)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "登录成功",
		Data: map[string]interface{}{
			"token":    token,
			"supplier": supplier,
		},
	})
}

// 供应商产品页
// 从该供应商的supply_list得到供应产品列表
// 从supplier_product_price中得到这个供应商的商品价格
func SupplierProducts(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var supSrv SupplierService
	if err := c.ShouldBindJSON(&supSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "param error",
		})
		return
	}

	collection := models.Client.Collection("supplier")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["supplier_id"] = supSrv.SupplierID

	var supplier models.Supplier
	err := collection.FindOne(context.TODO(), filter).Decode(&supplier)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "找不到供应商",
		})
		return
	}

	var responseData []SupplierProductInfo

	if len(supplier.SupplyList) > 0 {
		collection := models.Client.Collection("supplier_product_price")
		productCollection := models.Client.Collection("product")
		filter := bson.M{}
		filter["com_id"] = claims.ComId
		filter["product_id"] = bson.M{"$in": supplier.SupplyList}
		filter["supplier_id"] = supplier.ID
		filter["is_valid"] = true
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "找不到商品对应销售价",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var res models.SupplierProductPrice
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode error",
				})
				return
			}
			var product models.Product
			err := productCollection.FindOne(context.TODO(), bson.D{{"product_id", res.ProductID}}).Decode(&product)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode error",
				})
				return
			}
			supProInfo := SupplierProductInfo{
				ProudctPrice: res,
				Urls:         product.URLS,
			}
			responseData = append(responseData, supProInfo)
		}

		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "product list",
			Data: responseData,
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: -1,
		Msg:  "暂时没有供应的商品",
	})
}

// 供应商订单页面
// 待处理1 已发货2 已完成3 全部订单0 四个状态
func SupplierOrders(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var (
		supSrv   SupplierService
		orders   []models.SupplierOrder
		orderIds []int64
	)

	if err := c.ShouldBindJSON(&supSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "param error",
		})
		return
	}

	// 1 先在主表中找出该供应商对应状态的订单
	collection := models.Client.Collection("supplier_order")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	if supSrv.OrderStatus > 0 {
		filter["status"] = supSrv.OrderStatus // 待处理1
	}
	filter["supplier_id"] = supSrv.SupplierID

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "找不到采购订单",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.SupplierOrder
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode error",
			})
			return
		}
		orders = append(orders, result)
		orderIds = append(orderIds, result.OrderId)
	}

	// 根据主表的订单号找出子订单
	var supOrders []models.SupplierSubOrder
	if len(orderIds) > 0 {
		collection := models.Client.Collection("supplier_sub_order")
		filter := bson.M{}
		filter["com_id"] = claims.ComId
		filter["order_id"] = bson.M{"$in": orderIds}
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "找不到对应采购子订单",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var res models.SupplierSubOrder
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode error",
				})
				return
			}
			supOrders = append(supOrders, res)
		}
	}

	type respData struct {
		Orders    []models.SupplierOrder    `json:"orders"`
		SubOrders []models.SupplierSubOrder `json:"sub_orders"`
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get all supplier orders",
		Data: respData{
			Orders:    orders,
			SubOrders: supOrders,
		},
	})
}

// 供应商操作页面，填写该单的实发数量和上传凭证（图片）
// 获取数据的格式
// {
//   "order_id":"",
//   "actual_amount":"",
//   "urls": [],
// }

//1、未发货
//2、已收货
//3、已发货
//4、部分收货
//5，审核通过
//6，审核不通过
func SupplierExecOrder(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var supSrv SupplierService

	if err := c.ShouldBindJSON(&supSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "param error",
		})
		return
	}

	// 更新该单的实发数量
	// 更新订单的状态
	// 如果系统设置了发货误差率，则还要计算一下传回来的数据是否在误差范围之内
	collection := models.Client.Collection("supplier_sub_order")
	for _, sub_order := range supSrv.SubOrders {
		_, err := collection.UpdateOne(context.TODO(),
			bson.D{{"com_id", claims.ComId},
				{"order_sub_id", sub_order.SubOrderID}},
			bson.M{
				"$set": bson.M{"actual_amount": sub_order.ActualAmount,
								"state" : 3}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't update supplier sub order",
			})
			return
		}
	}

	// 更新总订单的状态
	collection = models.Client.Collection("supplier_order")
	_, err := collection.UpdateOne(context.TODO(), bson.D{{"order_id", supSrv.OrderID}}, bson.M{"$set": bson.M{"status": 3}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update supplier order",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "操作成功",
	})
}

// 供应商结算单页面
func SupplierSettlements(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var supSrv SupplierService
	if err := c.ShouldBindJSON(&supSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "param error",
		})
		return
	}

	// 根据settlement_type 来选择结算单
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["type"] = 2
	filter["SupSettle"] = supSrv.SettlementType

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find supplier settlement")
		return
	}

	var settlementList []models.GoodsInstance
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode goods instance: ", err)
			return
		}
		settlementList = append(settlementList, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "get settlemt list",
		Data: settlementList,
	})
}

func SupplierResetPassword(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var resetPassSrv SupplierResetPasswordService
	if err := c.ShouldBindJSON(&resetPassSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Param error",
		})
		return
	}

	if resetPassSrv.NewPassword != resetPassSrv.ConfirmPass {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "两次输入的密码不一样",
		})
		return
	}

	supplier := models.Supplier{}
	supplier.ComID = claims.ComId
	supResp, err := supplier.FindByID(resetPassSrv.SupplierID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get supplier",
		})
		return
	}
	if supResp.Password != resetPassSrv.OldPassword {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Old password wrong",
		})
		return
	}
	// TODO: SHA加密
	if err := supResp.UpdatePassword(resetPassSrv.NewPassword); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update password",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update password succeed",
	})

}

// 供应商上传凭证
func SupplierUploadCertificate(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 获取图片保存地址
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}

	files := form.File["file"]
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get store path",
		})
		return
	}

	save_path := dir
	urls := []string{}

	for _, file := range files {
		path, filename := util.GetYpyunPath(file.Filename)
		upload_path := save_path + path
		_, err = os.Stat(upload_path)

		if os.IsNotExist(err) {
			util.Log().Error("path not exist: ", err.Error())
			err = os.MkdirAll(upload_path, os.ModePerm)

			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't create filepath",
				})
				return
			}
		}
		err := c.SaveUploadedFile(file, upload_path+filename)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't save image",
			})
			return
		}
		ypyunURL1 := "/upload/" + strconv.Itoa(int(claims.ComId)) + "/product_img/" + filename
		err = util.UpYunPut(ypyunURL1, upload_path+filename)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't upload image",
			})
			return
		}
		ret_url := "http://img.jxc.weqi.exechina.com" + ypyunURL1
		urls = append(urls, ret_url)
	}

	order_id, err := strconv.Atoi(c.Query("order_id"))
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get supplier order",
		})
		return
	}

	// 更新订单凭证图片地址
	var supplierOrder *models.SupplierOrder
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = order_id
	update := bson.M{"$set": bson.M{"order_urls": urls}}
	err = supplierOrder.UpdateOneByOrderID(filter, update)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get supplier order",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get supplier instance",
		Data: urls,
	})
}
