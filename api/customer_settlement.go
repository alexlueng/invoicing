package api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/serializer"
	"net/http"
	"time"

	//"context"
	"encoding/json"
	//"fmt"
	"github.com/gin-gonic/gin"
	//"jxc/serializer"
	//"net/http"

	//"github.com/gin-gonic/gin/internal/json"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/models"
)

func ListCustomerSettlement(c *gin.Context) {
	//com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	//if err != nil || models.THIS_MODULE != int(com.ModuleId) {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "域名错误",
	//	})
	//	return
	//}

	var req models.CustomerSettlementReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &req)

	fmt.Println("View: ", req.View)

	var list []models.GoodsInstance



	// 找到订单实例表中的状态为确认的记录，显示出来
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}


	option := options.Find() // 按照req.View的值来进行排序

	switch req.View {
	case "customer":
		fmt.Println("customer view")
		// 先得到所有客户

		//result := make(map[string]map[string]interface{})
		result := make(map[string][]models.GoodsInstance)
		customer := models.Customer{}
		cusList, err := customer.FindAll(filter, option)
		if err != nil {
			fmt.Println("Can't find this customer: ", err)
			return
		}
		var cusArray []int64
		for _, cus := range cusList {
			cusArray = append(cusArray, cus.ID)
			//result[cus.Name] = make(map[string]interface{})
			result[cus.Name] = []models.GoodsInstance{}
			//result[cus.Name]["total_amount"] = 0
		}

		fmt.Println("the new map is: ", result)

		// 找出所有客户的商品实例
		filter["dest_id"] = bson.M{"$in": cusArray}
		filter["com_id"] = 1
		filter["type"] =bson.M{"$in": []int64{1, 3}} //类型 1是从仓库发向客户 3是从供应商发向客户
		filter["dest_type"] = 1 // 只查看去向客户的实例单
		filter["status"] = 3 //状态为3是已审核的订单，进入结算

		cur, err := collection.Find(context.TODO(), filter)

		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode good instance: ", err)
				return
			}
			result[res.DestTitle] = append(result[res.DestTitle], res)
			//result[res.DestTitle]["all_good_instances"] = append(result[res.DestTitle]["all_good_instances"].([]models.GoodsInstance), res)
			//resMap["all_good_instances"] = customerList
			//resMap["total_amount"] = totalAmount
			//resMap["un_settlement"] = unSettlement
			//resMap["settling"] = settling
			//resMap["settled"] = settled
			//resMap["lastest"] = "" //最近结算时间
			//result[res.DestTitle]["total_amount"] = res.CustomerPrice * float64(res.Amount)
			//switch res.Settlement {
			//case 1: // 未结算
			//	result[res.DestTitle]["un_settlement"] = result[res.DestTitle]["un_settlement"].(int64) + 1
			//case 2: // 结算中
			//	result[res.DestTitle]["settling"] = result[res.DestTitle]["settling"].(float64) + res.CustomerPrice * float64(res.Amount)
			//case 3: // 已结算
			//	result[res.DestTitle]["settled"] = result[res.DestTitle]["settled"].(float64) + res.CustomerPrice * float64(res.Amount)
			//default:
			//}
		}
		for

		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Get customer instance",
			Data: result,
		})
		return

	case "settlement":
	default:
		fmt.Println("exec here")
		fmt.Println("filter: ", filter)
		cur, err := collection.Find(context.TODO(), filter, option)
		if err != nil {
			fmt.Println("can't find cusOrderInstance: ", err)
			return
		}

		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode customer order instance: ", err)
				return
			}
			list = append(list, res)
		}
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Get customer instance",
			Data: list,
		})
		return

	}


	//if req.View == "" { // 默认视图，以订单号模式展示
	//	fmt.Println("exec here")
	//	fmt.Println("filter: ", filter)
	//	cur, err := collection.Find(context.TODO(), filter, option)
	//	if err != nil {
	//		fmt.Println("can't find cusOrderInstance: ", err)
	//		return
	//	}
	//
	//	for cur.Next(context.TODO()) {
	//		var res models.GoodsInstance
	//		if err := cur.Decode(&res); err != nil {
	//			fmt.Println("Can't decode customer order instance: ", err)
	//			return
	//		}
	//		list = append(list, res)
	//	}
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: 200,
	//		Msg:  "Get customer instance",
	//		Data: list,
	//	})
	//	return
	//
	//} else if req.View == "customer" {
	//	fmt.Println("customer view")
	//	// 先得到所有客户
	//	customer := models.Customer{}
	//	filter = bson.M{}
	//	option = options.Find()
	//	cusList, err := customer.FindAll(filter, option)
	//	if err != nil {
	//		fmt.Println("Can't find this customer: ", err)
	//		return
	//	}
	//	// 查找这个客户所有的订单实例
	//	for _, cus := range cusList {
	//	//	fmt.Printf("customer name: %s, customer id: %d\n", cus.Name, cus.ID)
	//		var totalAmount float64 // 总金额
	//		var unSettlement int64 //  未结算单数
	//		var settling float64 // 结算中的金额
	//		var settled float64 // 已结算金额
	//		var customerList []models.GoodsInstance
	//		filter := bson.M{}
	//		filter["dest_id"] = cus.ID
	//		cur, err := collection.Find(context.TODO(), filter)
	//		if err != nil {
	//			fmt.Println("Can't find good instances of this customer: ", err)
	//			return
	//		}
	//		for cur.Next(context.TODO()) {
	//			var instance models.GoodsInstance
	//			if err := cur.Decode(&instance); err != nil {
	//				fmt.Println("Can't decode into goodInstance: ", err)
	//				return
	//			}
	//			totalAmount += instance.CustomerPrice * float64(instance.Amount)
	//			switch instance.Settlement {
	//			case 1: // 未结算
	//				unSettlement += 1
	//			case 2: // 结算中
	//				settling += instance.CustomerPrice * float64(instance.Amount)
	//			case 3: // 已结算
	//				settled += instance.CustomerPrice * float64(instance.Amount)
	//			default:
	//			}
	//			customerList = append(customerList, instance)
	//		}
	//		resMap := make(map[string]interface{})
	//		resMap["all_good_instances"] = customerList
	//		resMap["total_amount"] = totalAmount
	//		resMap["un_settlement"] = unSettlement
	//		resMap["settling"] = settling
	//		resMap["settled"] = settled
	//		resMap["lastest"] = "" //最近结束时间
	//		result[cus.Name] = resMap
	//
	//	}
		// 未结算单数

		// 结算中
		// 已结算
		// 最近结算
		// 待结算
		// 总金额
	//}
	//total, _ := collection.CountDocuments(context.TODO(), filter)
	//
	//// 返回查询到的总数，总页数
	//resData := models.ResponseCustomerSettlementData{}
	//resData.CustomerOrderInstance = cusOrderInstances
	//resData.Total = int(total)
	//resData.Pages = int(total/req.Size) + 1
	//resData.Size = int(req.Size)
	//resData.CurrentPage = int(req.Page)
	//c.JSON(http.StatusOK, serializer.Response{
	//	Code: 200,
	//	Msg:  "Get customer instance",
	//	Data: result,
	//})

}

type GenSettlementService struct {
	CustomerName string `json:"customer_name"`
	CustomerID int64 `json:"customer_id"`
	settlementList []string `json:"settlement_list"`
}

// 生成结算单
func GenSettlement(c *gin.Context) {
	//com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	//if err != nil || models.THIS_MODULE != int(com.ModuleId) {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "域名错误",
	//	})
	//	return
	//}

	data, _ := ioutil.ReadAll(c.Request.Body)
	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, gs)

	SmartPrint(gs)

	collection := models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance
	for _, instance_sn := range gs.settlementList {
		var instance models.GoodsInstance
		filter := bson.M{}
		filter["com_id"] = 1
		filter["instance_sn"] = instance_sn
		err := collection.FindOne(context.TODO(), filter).Decode(&instance)
		if err != nil {
			fmt.Println("Can't found the good instance: ", err)
			return
		}
		instanceList = append(instanceList, instance)
	}

	cusSettlement := models.CustomerSettlement{}
	//得到一个自增ID
	//生成一个结算单号
	cusSettlement.ID = getLastID("cus_settlement")
	// 设置结算单状态
	cusSettlement.SettlementSN = GetTempOrderSN() // 需要一个独立的方法
	cusSettlement.CustomerName = gs.CustomerName
	cusSettlement.CustomerID = gs.CustomerID
	for _, instance := range instanceList {
		cusSettlement.CustomerInstance = append(cusSettlement.CustomerInstance, instance)
		cusSettlement.SettlementAmount += instance.CustomerPrice * float64(instance.Amount)
	}
	cusSettlement.UnpaidAmount = cusSettlement.SettlementAmount
	current_time := time.Now()
	cusSettlement.CreateTime = current_time.Unix()
	cusSettlement.Status = 1 // 1表示正在结算 2表示结算完成

	collection = models.Client.Collection("customer_settlement")
	_, err := collection.InsertOne(context.TODO(), cusSettlement)
	if err != nil {
		fmt.Println("Insert customer error: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Insert customer settlement succeed",
		//Data: result,
	})
}