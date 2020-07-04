package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 采购订单表结构 (供应商订单)

// 接收数据
// 供应商信息

// 商品数组

// 接收方信息

// 创建供应商订单
func CreateSupplierOrder(supplierOrder models.SupplierOrder, supplierOrderInstance []interface{}) (error) {
	// 添加供应商订单
	supplierOrderCollection := models.Client.Collection("supplier_order")
	_, err := supplierOrderCollection.InsertOne(context.TODO(), supplierOrder)
	if err != nil {
		return err
	}

	// 添加供应商订单商品实例
	supplierOrderInstanceCollection := models.Client.Collection("supplier_sub_order")
	_, err = supplierOrderInstanceCollection.InsertMany(context.TODO(), supplierOrderInstance)
	// 如果实例添加失败则删除对应的供应商订单
	if err != nil {
		supplierOrderCollection.DeleteOne(context.TODO(), bson.M{"order_sn": supplierOrder.OrderSN})
		return err
	}

	// 修改供应商交易次数
	collection := models.Client.Collection("supplier")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"supplier_id",supplierOrder.SupplierID}, {"com_id",supplierOrder.ComID}}, bson.M{
		"$inc" : bson.M{"transaction_num" : 1}})
	if err != nil {
		fmt.Println("Can't update supplier transaction num: ", err)
		return err
	}
	fmt.Println("update result: ", updateResult.UpsertedID)

	return nil

}

//
