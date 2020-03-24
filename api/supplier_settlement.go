package api

import (
	"context"
	"fmt"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"time"
)

type SupSettlementResponseDetail struct {
	InstanceList []models.GoodsInstance
	UnSettleAmount int64
	TotalAmount float64
	UnSettlement float64
	Settled float64
	Settling float64
	LastSettle int64
	SupplierName string
	SupplierID int64
}

func ListSupplierSettlement(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.SupplierSettlementReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &req)

	fmt.Println("View: ", req.View)

	option := options.Find() // 按照req.View的值来进行排序
	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderField := []string{"OrderSN", "price"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "OrderSN"
	}
	// 设置排序顺序 desc asc
	order := 1
	fmt.Println("order: ", req.Ord)
	if req.Ord == "desc" {
		order = -1
		req.Ord = "desc"
	} else {
		order = 1
		req.Ord = "asc"
	}

	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
	option.SetSort(bson.D{{req.OrdF, order}})

	collection := models.Client.Collection("goods_instance")
	resData := models.ResponseSupplierSettlementData{}
	var total int64
	filter := bson.M{}

	switch req.View {
	//供应商视图
	case "supplier":
		fmt.Println("supplier view")
		// 先得到所有供应商

		//result := make(map[string]map[string]interface{})
		SupplierSettlements := make(map[string][]models.GoodsInstance)


		collection := models.Client.Collection("goods_instance")

		supplierIDS, err := collection.Distinct(context.TODO(), "src_id", bson.M{"status":bson.M{"$eq":3}, "src_type":bson.M{"$eq":2}})
		if err != nil {
			fmt.Println("can't distinct customer: ", err)
			return
		}
		fmt.Println(supplierIDS)

		total = int64(len(supplierIDS))

		// 找出所有供应商的商品实例
		filter["src_id"] = bson.M{"$in": supplierIDS}
		filter["com_id"] = claims.ComId
		//filter["src_type"] =bson.M{"$in": []int64{2, 3}} //类型 1是从仓库发向客户 3是从供应商发向客户
		filter["src_type"] = 2 // 只查看来源供应商的实例单
		filter["status"] = 3 //状态为3是已审核的订单，进入结算

		cur, err := collection.Find(context.TODO(), filter)
		total, _ = collection.CountDocuments(context.TODO(), filter)

		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode good instance: ", err)
				return
			}
			SupplierSettlements[res.SrcTitle] = append(SupplierSettlements[res.SrcTitle], res)

		}

		var result []SupSettlementResponseDetail

		for k, v := range SupplierSettlements {
			detail := SupSettlementResponseDetail{}
			detail.InstanceList = v
			detail.SupplierName = k
			detail.SupplierID = v[0].SrcId

			fmt.Println("Supplier id is: ", detail.SupplierID)

			sup := models.Supplier{}
			sup.ComID = claims.ComId
			supRes, err := sup.FindByID(detail.SupplierID)
			fmt.Println(supRes)
			if err != nil {
				fmt.Println("Can't find customer: ", err)
				return
			}
			detail.LastSettle = supRes.LastSettlement

			for _, ins := range v {
				detail.TotalAmount += ins.SupplierPrice * float64(ins.Amount) // 总金额
				switch ins.SupSettle {
				case 0:
					detail.UnSettleAmount += 1 // 未结算单数
					detail.UnSettlement += ins.SupplierPrice * float64(ins.Amount) // 未结算金额
				case 1:
					detail.Settling += ins.SupplierPrice * float64(ins.Amount)
				case 2:
					detail.Settled += ins.SupplierPrice * float64(ins.Amount) // 已结算金额

				}

			}

			result = append(result, detail)
		}


		resData.Result = result

	// 结算单视图
	case "settlement":
		fmt.Println("settlement view")
		supSettlementCollects := models.Client.Collection("supplier_settlement")
		var result []models.SupplierSettlement
		cur, err := supSettlementCollects.Find(context.TODO(), bson.D{{"com_id", claims.ComId}}, option)
		total, _ = supSettlementCollects.CountDocuments(context.TODO(), bson.D{{"com_id", claims.ComId}})
		if err != nil {
			fmt.Println("Can't get supplier settlement: ", err)
			return
		}
		for cur.Next(context.TODO()) {
			var res models.SupplierSettlement
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode supplier settlement: ", err)
				return
			}
			result = append(result, res)
		}

		resData.Result = result

	// 默认视图
	default:
		fmt.Println("exec here")
		//fmt.Println("filter: ", filter)



		filter["com_id"] = claims.ComId
		filter["src_type"] = 2
		filter["status"] = 3

		var result []models.GoodsInstance

		cur, err := collection.Find(context.TODO(), filter, option)
		total, _ = collection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println("can't find supOrderInstance: ", err)
			return
		}

		for cur.Next(context.TODO()) {
			var res models.GoodsInstance
			if err := cur.Decode(&res); err != nil {
				fmt.Println("Can't decode Supplier order instance: ", err)
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
		Msg:  "Get supplier instance",
		Data: resData,
	})

}

type GenSupSettlementService struct {
	SettleTime int64 `json:"time"` // 选择结算日期
	SupplierName string `json:"supplier_name"`
	SupplierID int64 `json:"supplier_id"`
	InstanceIDs []int64 `json:"id_list"` // 商品实例id
	SettleType string `json:"type"` // 查询结算类型
}

// 生成结算单：
// 1，找出所有需要结算的商品实例 2，将实例中的settlement字段置为1 3，生成结算单并将记录插入到客户结算表中

func GenSupSettlement(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)


	// 生成结算单是按照两种方式 1、按照选定时间之前的所有未结算的商品实例 2、选择特定的商品实例进行结算

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSupSettlementService
	_ = json.Unmarshal(data, &gs)

	SmartPrint(gs)

	collection := models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance
	var instanceIDs []int64
	filter := bson.M{}

	// 找出所有符合条件的实例
	if gs.SupplierID != 0 && gs.SettleTime > 0 { // 按天数结算
		fmt.Println("according to supplier id")
		filter["confirm_time"] = bson.M{"$lt": gs.SettleTime}
	} else if len(gs.InstanceIDs) > 0 { // 按商品实例id
		fmt.Println("according to supplier instance")
		filter["instance_id"] = bson.M{"$in": gs.InstanceIDs}
	}

	filter["com_id"] = 1
	//filter["dest_id"] = gs.CustomerID
	filter["supsettle"] = 0 // 未结算
	filter["status"] = 3 // 审核过

	fmt.Println("filter: ", filter)

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find supplier instance: ", err)
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
	var supSettlement models.SupplierSettlement
	supSettlement.ComID = claims.ComId
	supSettlement.ID = getLastID("supplier_settlement")
	supSettlement.SupplierName = gs.SupplierName
	supSettlement.SettlementSN = GetTempOrderSN()
	supSettlement.SupplierID = gs.SupplierID

	supSettlement.SupplierInstance = instanceIDs

	current_time := time.Now()
	supSettlement.CreateTime = current_time.Unix()

	for _, ins := range instanceList {
		supSettlement.SettlementAmount += ins.SupplierPrice * float64(ins.Amount)
	}

	supSettlement.Status = 0 // 0:结算中 1：结算完成

	supSettleCollects := models.Client.Collection("supplier_settlement")
	insertResult, err := supSettleCollects.InsertOne(context.TODO(), supSettlement)
	if err != nil {
		fmt.Println("Can't insert cus settlement: ", err)
		return
	}
	fmt.Println("Insert result: ", insertResult.InsertedID)

	// 修改商品实例的状态
	updateResult, err := collection.UpdateMany(context.TODO(),
		bson.M{"instance_id": bson.M{"$in": instanceIDs}},
		bson.M{"$set": bson.M{ "supsettle" : 1,
			"supsettle_order_sn": supSettlement.SettlementSN,
			"supsettle_order_id": supSettlement.ID,}})
	if err != nil {
		fmt.Println("update many instance err: ", err)
		return
	}
	fmt.Println("update result: ", updateResult.UpsertedID)


	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "生成结算单成功",
	})
	//c.JSON(http.StatusOK, serializer.Response{
	//	Code: -1,
	//	Msg:  "生成结算单失败",
	//})

}


func FindSettlementSuppliers(c *gin.Context) {
	// 找出所有下过订单的供应商

	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)
	//TODO：带上comID去查
	// 支持 搜索，分页

	// 在订单列表中找
	collection := models.Client.Collection("goods_instance")

	supplierIDS, err := collection.Distinct(context.TODO(), "src_id", bson.M{"status":bson.M{"$eq":3}, "src_type":bson.M{"$eq":2}})
	if err != nil {
		fmt.Println("can't distinct supplier: ", err)
		return
	}
	fmt.Println(supplierIDS)
	collection = models.Client.Collection("supplier")
	cur, err := collection.Find(context.TODO(), bson.M{"supplier_id": bson.M{"$in": supplierIDS}})
	if err != nil {
		fmt.Println("error found: ", err)
		return
	}
	var result []models.Supplier
	for cur.Next(context.TODO()) {
		var res models.Supplier
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Error while decoding supplier: ", err)
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

func FindOneSupSettlements(c *gin.Context) {

	var req models.SupplierSettlementReq


	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSupSettlementService
	_ = json.Unmarshal(data, &gs)

	collection := models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance

	resData := models.ResponseSupplierSettlementData{}

	option := options.Find() // 按照req.View的值来进行排序
	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))

	filter := bson.M{}

	filter["com_id"] = claims.ComId
	filter["src_id"] = gs.SupplierID
	filter["src_type"] = 2
	filter["status"] = 3 // 审核过

	if gs.SettleTime > 0 { // 按时间选择
		filter["supsettle"] = 0 // 未结算
		filter["confirm_time"] = bson.M{"$lt": gs.SettleTime}

		fmt.Println("filter: ", filter)

		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			fmt.Println("Can't find supplier instance: ", err)
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
			Msg:  "Find supplier's settlement",
			Data: resData,
		})
		return
	}

	if gs.SettleType != "" {
		switch gs.SettleType {
		case "unsettle":
			fmt.Println("unsettle")
			filter["supsettle"] = 0
		case "settling":
			fmt.Println("settling")
			filter["supsettle"] = 1
		case "settled":
			fmt.Println("settled")
			filter["supsettle"] = 2
		}
		fmt.Println("filter here: ", filter)
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			fmt.Println("Can't find supplier instance: ", err)
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
			Msg:  "Find supplier's settlement",
			Data: resData,
		})

		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: -1,
		Msg:  "Can't find supplier's settlement",
	})

}

func SupSettlementDetail(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	collection := models.Client.Collection("supplier_settlement")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["id"] = gs.SettlementID
	var settlement models.SupplierSettlement
	err := collection.FindOne(context.TODO(), filter).Decode(&settlement)
	if err != nil {
		fmt.Println("can't find settlement: ", err)
		return
	}

	collection = models.Client.Collection("goods_instance")

	var instanceList []models.GoodsInstance
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["instance_id"] = bson.M{"$in": settlement.SupplierInstance}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find supplier instance: ", err)
	}
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode good instance: ",err)
		}
		instanceList = append(instanceList, res)
		//		instanceIDs = append(instanceIDs, res.InstanceId)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Find supplier's settlement",
		Data: instanceList,
	})
}

func SupSettlementConfirm(c *gin.Context) {
	// 获取token，解析token获取登录用户信息
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	// 接收要结算的商品实例单号
	var gs GenSettlementService
	_ = json.Unmarshal(data, &gs)

	fmt.Println("settlement id: ", gs.SettlementID)

	collection := models.Client.Collection("supplier_settlement")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["id"] = gs.SettlementID
	var settlement models.SupplierSettlement
	err := collection.FindOne(context.TODO(), filter).Decode(&settlement)
	if err != nil {
		fmt.Println("Can't find supplier settlement: ", err)
		return
	}
	//fmt.Println(settlement)
	confirmTime := time.Now().Unix()
	updateResult, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set" : bson.M{
			"finish_time": confirmTime,
			"status" : 1}})
	if err != nil {
		fmt.Println("Can't update supplierment: ", err)
		return
	}
	fmt.Println("Update supplier: ", updateResult.UpsertedID)



	insCollects := models.Client.Collection("goods_instance")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	fmt.Println("istance id: ", settlement.SupplierInstance)
	filter["instance_id"] = bson.M{"$in": settlement.SupplierInstance}

	updateResult, err = insCollects.UpdateMany(context.TODO(),
		bson.M{"instance_id": bson.M{"$in": settlement.SupplierInstance}},
		bson.M{"$set": bson.M{ "supsettle" : 2,}})

	//updateResult, err = insCollects.UpdateMany(context.TODO(), filter, bson.M{ "$set": bson.M{ "settlement" : 2} })

	if err != nil {
		fmt.Println("Can't update instances: ", err)
		return
	}
	fmt.Println("Update supplier: ", updateResult.UpsertedID)

	fmt.Println("supplier id: ", settlement.SupplierID)

	// 修改用户最后结算时间
	cusCollections := models.Client.Collection("supplier")
	updateResult, err = cusCollections.UpdateOne(context.TODO(),
		bson.M{"supplier_id": settlement.SupplierID},
		bson.M{"$set": bson.M{ "last_settlement" : time.Now().Unix(),}})
	if err != nil {
		fmt.Println("update supplier last settlement err: ", err)
		return
	}
	fmt.Println("Update supplier: ", updateResult.UpsertedID)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Find supplier's settlement",
	})

}










