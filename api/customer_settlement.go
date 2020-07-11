package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type ResponseDetail struct {
	InstanceList []models.GoodsInstance
	UnSettleAmount int64
	TotalAmount float64
	UnSettlement float64
	Settled float64
	Settling float64
	LastSettle int64
	CustomerName string
	CustomerID int64
}

func ListCustomerSettlement(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.CustomerSettlementReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &req)

	page, size := SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderFields := []string{"OrderSN", "price"}
	option := SetPaginationAndOrder(req.OrdF, orderFields, req.Ord, page, size)

	collection := models.Client.Collection("goods_instance")
	resData := models.ResponseCustomerSettlementData{}
	var total int64
	filter := bson.M{}

	switch req.View {
	//客户视图
	case "customer":
		fmt.Println("customer view")
		// 先得到所有客户
		CustomerSettlements := make(map[string][]models.GoodsInstance)

		collection := models.Client.Collection("goods_instance")

		customerIDS, err := collection.Distinct(context.TODO(), "dest_id", bson.M{"status":bson.M{"$eq":3}, "dest_type":bson.M{"$eq":1}, "com_id": claims.ComId})
		if err != nil {
			fmt.Println("can't distinct customer: ", err)
			return
		}
		fmt.Println(customerIDS)

		total = int64(len(customerIDS))

		// 找出所有客户的商品实例
		filter["dest_id"] = bson.M{"$in": customerIDS}
		filter["com_id"] = claims.ComId
		//filter["dest_type"] =1 //类型 1是从仓库发向客户 3是从供应商发向客户
		filter["dest_type"] = 1 // 只查看去向客户的实例单
		filter["status"] = 3 //状态为3是已审核的订单，进入结算

		cur, err := collection.Find(context.TODO(), filter)

		if err != nil {
			fmt.Println("Can't find customer instances: ", err)
			return
		}

		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode good instance: ", err)
				return
			}
			CustomerSettlements[res.DestTitle] = append(CustomerSettlements[res.DestTitle], res)

		}

		var result []ResponseDetail

		for k, v := range CustomerSettlements {
			detail := ResponseDetail{}
			detail.InstanceList = v
			detail.CustomerName = k
			detail.CustomerID = v[0].DestId

			fmt.Println("Customer id is: ", detail.CustomerID)

			cus := models.Customer{}
			cus.ComID = claims.ComId
			cusRes, err := cus.FindByID(detail.CustomerID)
			fmt.Println(cusRes)
			if err != nil {
				fmt.Println("Can't find customer: ", err)
				return
			}
			detail.LastSettle = cusRes.LastSettlement

			for _, ins := range v {
				detail.TotalAmount += ins.CustomerPrice * float64(ins.Amount) // 总金额
				switch ins.CusSettle {
				case 0:
					detail.UnSettleAmount += 1 // 未结算单数
					detail.UnSettlement += ins.CustomerPrice * float64(ins.Amount) // 未结算金额
				case 1:
					detail.Settling += ins.CustomerPrice * float64(ins.Amount) // 结算中
				case 2:
					detail.Settled += ins.CustomerPrice * float64(ins.Amount) // 已结算金额

				}

			}

			result = append(result, detail)
		}

		resData.Result = result

	// 结算单视图
	case "settlement":
		fmt.Println("settlement view")
		cusSettlementCollects := models.Client.Collection("customer_settlement")
		var result []models.CustomerSettlement
		cur, err := cusSettlementCollects.Find(context.TODO(), bson.D{{"com_id", claims.ComId}}, option)
		total, _ = cusSettlementCollects.CountDocuments(context.TODO(), bson.D{{"com_id", claims.ComId}})
		if err != nil {
			fmt.Println("Can't get customer settlement: ", err)
			return
		}
		for cur.Next(context.TODO()) {
			var res models.CustomerSettlement
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode customer settlement: ", err)
				return
			}
			result = append(result, res)
		}

		resData.Result = result

	// 默认视图，列出结算的订单实例
	default:
		fmt.Println("exec here")

		filter["com_id"] = claims.ComId
		filter["dest_type"] = 1
		filter["status"] = 3

		var result []models.GoodsInstance

		cur, err := collection.Find(context.TODO(), filter, option)
		total, _ = collection.CountDocuments(context.TODO(), filter)
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
			result = append(result, res)
		}
		resData.Result = result
	}

	// 返回查询到的总数，总页数

	resData.Total = int(total)
	resData.Pages = int(total/req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customer instance",
		Data: resData,
	})

}

type GenSettlementService struct {
	SettleTime int64 `json:"time"` // 选择结算日期
	CustomerName string `json:"customer_name"`
	CustomerID int64 `json:"customer_id"`
	InstanceIDs []int64 `json:"id_list"` // 商品实例id
	SettleType string `json:"type"` // 查询结算类型
	SettlementID int64 `json:"settlement_id"`
}

// 生成结算单：
// 1，找出所有需要结算的商品实例 2，将实例中的settlement字段置为1 3，生成结算单并将记录插入到客户结算表中

func GenSettlement(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 生成结算单是按照两种方式 1、按照选定时间之前的所有未结算的商品实例 2、选择特定的商品实例进行结算

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	SmartPrint(gs)

	collection := models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance
	var instanceIDs []int64
	filter := bson.M{}

	// 找出所有符合条件的实例
	if gs.CustomerID != 0 && gs.SettleTime > 0 { // 按天数结算
		fmt.Println("according to customer id")
		filter["confirm_time"] = bson.M{"$lt": gs.SettleTime}
	} else if len(gs.InstanceIDs) > 0 { // 按商品实例id
		fmt.Println("according to customer instance")
		filter["instance_id"] = bson.M{"$in": gs.InstanceIDs}
	}

	filter["com_id"] = claims.ComId
	//filter["dest_id"] = gs.CustomerID
	filter["cussettle"] = 0 // 未结算
	filter["status"] = 3 // 审核过

	fmt.Println("filter: ", filter)

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find customer instance: ", err)
	}
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode good instance: ",err)
		}
		instanceList = append(instanceList, res)
		instanceIDs = append(instanceIDs, res.InstanceId)
	}

	fmt.Println("instanceList: ", instanceList)

	fmt.Println("Instance ids: ", instanceIDs)

	// 生成结算单
	var cusSettlement models.CustomerSettlement
	cusSettlement.ComID = claims.ComId
	cusSettlement.ID = GetLastID("customer_settlement")
	cusSettlement.CustomerName = gs.CustomerName
	cusSettlement.SettlementSN = GetTempOrderSN()

	if gs.CustomerID == 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "User ID 不能为空",
		})
		return
	}

	cusSettlement.CustomerID = gs.CustomerID
	cusSettlement.CustomerInstance = instanceIDs
	cusSettlement.CreateTime = time.Now().Unix()

	for _, ins := range instanceList {
		cusSettlement.SettlementAmount += ins.CustomerPrice * float64(ins.Amount)
	}

	cusSettlement.Status = 0 // 0:结算中 1：结算完成

	cusSettleCollects := models.Client.Collection("customer_settlement")
	insertResult, err := cusSettleCollects.InsertOne(context.TODO(), cusSettlement)
	if err != nil {
		fmt.Println("Can't insert cus settlement: ", err)
		return
	}
	fmt.Println("Insert result: ", insertResult.InsertedID)

	SetLastID("customer_settlement")

	// 修改商品实例的状态
	_, err = collection.UpdateMany(context.TODO(),
		bson.M{"instance_id": bson.M{"$in": instanceIDs}},
		bson.M{"$set": bson.M{ "cussettle" : 1,
			"cussettle_order_sn": cusSettlement.SettlementSN,
			"cussettle_order_id": cusSettlement.ID, }})
	if err != nil {
		fmt.Println("update many instance err: ", err)
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "生成结算单成功",
	})
}


func FindSettlementCustomers(c *gin.Context) {
	// 找出所有下过订单的客户
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	// TODO: 要带上comID去查询
	// TODO:支持 搜索，分页
	// 在订单列表中找
	collection := models.Client.Collection("goods_instance")

	// mongodb去重查询
	customerIDS, err := collection.Distinct(context.TODO(), "dest_id", bson.M{"com_id": claims.ComId, "status":bson.M{"$eq":3}, "dest_type":bson.M{"$eq":1}})
	if err != nil {
		fmt.Println("can't distinct customer: ", err)
		return
	}
	fmt.Println(customerIDS)
	collection = models.Client.Collection("customer")
	cur, err := collection.Find(context.TODO(), bson.M{"customer_id": bson.M{"$in": customerIDS}})
	if err != nil {
		fmt.Println("error found: ", err)
		return
	}
	var result []models.Customer
	for cur.Next(context.TODO()) {
		var res models.Customer
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Error while decoding customer: ", err)
			return
		}
		result = append(result, res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "get customer",
		Data: result,
	})
}

func FindOneSettlements(c *gin.Context) {

	var req models.CustomerSettlementReq

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	collection := models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance

	resData := models.ResponseCustomerSettlementData{}

	option := options.Find() // 按照req.View的值来进行排序
	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
//	var instanceIDs []int64
	filter := bson.M{}

	filter["com_id"] = claims.ComId
	filter["dest_id"] = gs.CustomerID
	filter["dest_type"] = 1
	filter["status"] = 3 // 审核过

	if gs.SettleTime > 0 { // 按时间选择
		filter["cussettle"] = 0 // 未结算
		filter["confirm_time"] = bson.M{"$lt": gs.SettleTime}

		fmt.Println("filter: ", filter)

		cur, err := collection.Find(context.TODO(), filter, option)
		if err != nil {
			fmt.Println("Can't find customer instance: ", err)
		}
		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode good instance: ",err)
			}
			instanceList = append(instanceList, res)
		}

		total, _ := collection.CountDocuments(context.TODO(), filter)
		// 返回查询到的总数，总页数
		resData.Result = instanceList
		resData.Total = int(total)
		resData.Pages = int(total/req.Size) + 1
		resData.Size = int(req.Size)
		resData.CurrentPage = int(req.Page)

		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Find customer's settlement",
			Data: resData,
		})
		return
	}

	if gs.SettleType != "" {
		switch gs.SettleType {
		case "unsettle":
			fmt.Println("unsettle")
			filter["cussettle"] = 0
		case "settling":
			fmt.Println("settling")
			filter["cussettle"] = 1
		case "settled":
			fmt.Println("settled")
			filter["cussettle"] = 2
		}
		fmt.Println("filter here: ", filter)
		cur, err := collection.Find(context.TODO(), filter, option)
		if err != nil {
			fmt.Println("Can't find customer instance: ", err)
		}
		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode good instance: ",err)
			}
			instanceList = append(instanceList, res)
			//		instanceIDs = append(instanceIDs, res.InstanceId)
		}

		total, _ := collection.CountDocuments(context.TODO(), filter)
		// 返回查询到的总数，总页数
		resData.Result = instanceList
		resData.Total = int(total)
		resData.Pages = int(total/req.Size) + 1
		resData.Size = int(req.Size)
		resData.CurrentPage = int(req.Page)

		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Find customer's settlement",
			Data: resData,
		})
		return
	}


	c.JSON(http.StatusOK, serializer.Response{
		Code: -1,
		Msg:  "Can't find customer's settlement",
	})

}

// 结算单详情（这个可以放到客户结算单请求列表那里，像客户订单一样）
func SettlementDetail(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	collection := models.Client.Collection("customer_settlement")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["id"] = gs.SettlementID
	var settlement models.CustomerSettlement
	err := collection.FindOne(context.TODO(), filter).Decode(&settlement)
	if err != nil {
		fmt.Println("can't find settlement: ", err)
		return
	}

	collection = models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	if len(settlement.CustomerInstance) == 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "No customer instances found",
		})
	}
	filter["instance_id"] = bson.M{"$in": settlement.CustomerInstance}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find customer instance: ", err)
	}
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode good instance: ",err)
			return
		}
		instanceList = append(instanceList, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Find customer's settlement",
		Data: instanceList,
	})
}

// 客户确认结算
func SettlementConfirm(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	fmt.Println("settlement id: ", gs.SettlementID)

	collection := models.Client.Collection("customer_settlement")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["id"] = gs.SettlementID
	//
	var settlement models.CustomerSettlement
	err := collection.FindOne(context.TODO(), filter).Decode(&settlement)
	if err != nil {
		fmt.Println("Can't find customer settlement: ", err)
		return
	}
	// 更新结算单的状态
	confirmTime := time.Now().Unix()
	updateResult, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set" : bson.M{
			"finish_time": confirmTime,
			"status" : 1}})
	if err != nil {
		fmt.Println("Can't update customerment: ", err)
		return
	}
	fmt.Println("Update customer: ", updateResult.UpsertedID)
	// 更新结算单中商品实例的状态
	insCollects := models.Client.Collection("goods_instance")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	fmt.Println("istance id: ", settlement.CustomerInstance)
	filter["instance_id"] = bson.M{"$in": settlement.CustomerInstance}

	updateResult, err = insCollects.UpdateMany(context.TODO(),
		bson.M{"instance_id": bson.M{"$in": settlement.CustomerInstance}},
		bson.M{"$set": bson.M{ "cussettle" : 2,}})

	if err != nil {
		fmt.Println("Can't update instances: ", err)
		return
	}

	fmt.Println("update result: ", updateResult.UpsertedID)

	fmt.Println("customer id: ", settlement.CustomerID)

	// 修改用户最后结算时间
	// 更新客户管理页面的已付金额
	cusCollections := models.Client.Collection("customer")
	updateResult, err = cusCollections.UpdateOne(context.TODO(),
		bson.M{"customer_id": settlement.CustomerID},
		bson.M{"$set": bson.M{ "last_settlement" : time.Now().Unix()}})
	if err != nil {
		fmt.Println("update customer last settlement err: ", err)
		return
	}
	fmt.Println("Update customer: ", updateResult.UpsertedID)

	// 更新客户管理页面的已付金额
	fmt.Println("customers paid amount: ", settlement.SettlementAmount)
	updateResult, err = cusCollections.UpdateOne(context.TODO(),
		bson.M{"customer_id": settlement.CustomerID},
		bson.M{"$inc": bson.M{ "paid" : settlement.SettlementAmount}})
	if err != nil {
		fmt.Println("update customer paid amount err: ", err)
		return
	}
	fmt.Println("Update customer: ", updateResult.UpsertedID)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Find customer's settlement",
		Data: settlement,
	})

}

type DownloadService struct {
	DestID int64 `json:"dest_id"`
}

// 将商品实例导出Excel
func CustomerSettlementDownload(c *gin.Context) {

	// 根据域名得到COMID
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var id DownloadService
	if err := c.ShouldBindJSON(&id); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "生成结算单失败",
		})
	}

	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["dest_id"] = id.DestID


	collection := models.Client.Collection("goods_instance")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't get instance: ", err)
		return
	}
	var instances []models.GoodsInstance

	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		err := cur.Decode(&res)
		if err != nil {
			fmt.Println("Can't get instance: ", err)
			return
		}
		instances = append(instances, res)
	}

	for _, item := range instances {
		fmt.Println(item)
	}

	xlsx := excelize.NewFile()

	xlsx.MergeCell("Sheet1", "A1", "H1")
	xlsx.SetRowHeight("Sheet1", 1, 40)

	xlsx.SetColWidth("Sheet1", "F", "F", 30)
	xlsx.SetColWidth("Sheet1", "G", "G", 30)
	//设置单元格样式
	//"fill":{"type":"pattern","color":["#CCFFFF"],
	//"pattern":1}}
	style, err := xlsx.NewStyle(`{"alignment":{"horizontal":"center","Vertical":"center"},
				"font":{"bold":true, "size": 30},
				"border":[{"type":"left","color":"FF0000","style":1}]}`)
	if err != nil {
		fmt.Println(err.Error())
	}

	xlsx.SetCellStyle("Sheet1", "A1", "A1", style)
	xlsx.SetCellValue("Sheet1", "A1", instances[0].DestTitle + time.Now().Format("2006-01-02") + "结算单")

	//xlsx.SetCellValue("Sheet1", "A2", "结算单号")
	xlsx.SetCellValue("Sheet1", "A2", "商品名称")
	xlsx.SetCellValue("Sheet1", "B2", "联系人")
	xlsx.SetCellValue("Sheet1", "C2", "售价")
	xlsx.SetCellValue("Sheet1", "D2", "数量")
	xlsx.SetCellValue("Sheet1", "E2", "总金额")
	xlsx.SetCellValue("Sheet1", "F2", "下单时间")
	xlsx.SetCellValue("Sheet1", "G2", "发货时间")
	xlsx.SetCellValue("Sheet1", "H2", "确认状态")

	var totalPrice float64
	var totalAmount int64

	for i, item := range instances {

		lineNo := i + 3
		strLineNo := strconv.Itoa(lineNo)

		//xlsx.SetCellValue("Sheet1", "A" + strLineNo, item.CusSettleOrderSN)
		xlsx.SetCellValue("Sheet1", "A" + strLineNo, item.Product)
		xlsx.SetCellValue("Sheet1", "B" + strLineNo, item.Contacts)
		xlsx.SetCellValue("Sheet1", "C" + strLineNo, item.CustomerPrice)
		xlsx.SetCellValue("Sheet1", "D" + strLineNo, item.Amount)
		xlsx.SetCellValue("Sheet1", "E" + strLineNo, item.CustomerPrice * float64(item.Amount))
		xlsx.SetCellValue("Sheet1", "F" + strLineNo, time.Unix(item.OrderTime, 0).Format("2006-01-02 15:04:05")) // time.Unix(timestamp, 0).Format(timeLayout)
		xlsx.SetCellValue("Sheet1", "G" + strLineNo, time.Unix(item.ShipTime, 0).Format("2006-01-02 15:04:05"))
		if item.Status == 3 {
			xlsx.SetCellValue("Sheet1", "H" + strLineNo, "已确认")
			totalAmount += item.Amount
			totalPrice += item.CustomerPrice * float64(item.Amount)
		} else {
			xlsx.SetCellValue("Sheet1", "H" + strLineNo, "未确认，不计入本次结算")
		}
	}
	endLine := strconv.Itoa(len(instances) + 3)

	xlsx.SetCellValue("Sheet1", "C" + endLine, "总计")
	xlsx.SetCellValue("Sheet1", "D" + endLine, totalAmount)
	xlsx.SetCellValue("Sheet1", "E" + endLine, totalPrice)


	//保存文件方式
	//_ = xlsx.SaveAs("./aaa.xlsx")

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=" + strconv.Itoa(int(id.DestID)) + ".xlsx")
	c.Header("Content-Transfer-Encoding", "binary")

	//回写到web 流媒体 形成下载
	_ = xlsx.Write(c.Writer)

}





