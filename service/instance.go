package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 获取子订单实例列表
func FindSubOrderInstance(subOrderId int64, comId int64) ([]models.GoodsInstance, error) {
	collection := models.Client.Collection("goods_instance")
	var instance models.GoodsInstance
	var instanceArr []models.GoodsInstance
	// 搜索条件：发给客户 dest_type = 1,子订单id
	filter := bson.M{}
	filter["com_id"] = comId
	filter["dest_type"] = 1
	filter["dest_sub_order_id"] = subOrderId
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&instance)
		if err != nil {
			return nil, err
		}
		instanceArr = append(instanceArr, instance)
	}
	return instanceArr, nil
}

// 查找实例
func FindOneInstance(instanceId int64, comId int64) (*models.GoodsInstance, error) {
	collection := models.Client.Collection("goods_instance")
	var instance models.GoodsInstance
	filter := bson.M{}
	filter["com_id"] = comId
	filter["instance_id"] = instanceId
	err := collection.FindOne(context.TODO(), filter).Decode(&instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// 更新实例信息
func UpdateInstance(instanceId int64, update map[string]string, comId int64) error {
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = comId
	filter["instance_id"] = instanceId
	var updateBson bson.M
	for key, val := range update {
		updateBson[key] = val
	}
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": updateBson})
	if err != nil {
		return err
	}
	return nil
}

// 根据采购子订单号获取实例
// 来源为供应商 src_type = 2，src_sub_order_sn = sub_order_sn
func SubOrderSnFindOneInstance(subOrderSn string, comId int64) (*models.GoodsInstance, error) {
	collection := models.Client.Collection("goods_instance")
	var instance models.GoodsInstance
	filter := bson.M{}
	filter["com_id"] = comId
	filter["src_type"] = 2
	filter["src_sub_order_sn"] = subOrderSn
	err := collection.FindOne(context.TODO(), filter).Decode(&instance)
	if err != nil {
		return nil, err
	}
	return &instance, err
}
