package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 仓库相关方法

// 查找仓库
func FindOneWarehouse(warehouse_id, com_id int64) (*models.Warehouse, error) {
	collection := models.Client.Collection("warehouse")
	var warehouse models.Warehouse
	filter := bson.M{}
	filter["warehouse_id"] = warehouse_id
	filter["com_id"] = com_id

	err := collection.FindOne(context.TODO(), filter).Decode(&warehouse)
	if err != nil {
		return nil, err
	}
	// 获取仓库职员
	stuffs, err := FindOneWarehouseStuffs(warehouse_id, com_id)
	if err != nil {
		return nil, err
	}
	warehouse.WarehouseStuff = stuffs
	return &warehouse, nil
}

// 如果没有提交仓库id则获取公司所有仓库信息
// 如果有则获取所需部分
func FindWarehouse(com_id int64) (map[int64]models.Warehouse, error) {
	collection := models.Client.Collection("warehouse")
	var warehouse models.Warehouse
	warehouseArr := make(map[int64]models.Warehouse) // map[Warehouse_id]models.Warehouse

	filter := bson.M{}
	//filter["warehouse_id"] = warehouse_id
	filter["com_id"] = com_id

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&warehouse)
		if err != nil {
			return nil, err
		}
		warehouseArr[warehouse.ID] = warehouse
	}
	return warehouseArr, nil
}

// 查找某仓库的职员
func FindOneWarehouseStuffs(warehouse_id int64, com_id int64) ([]models.WarehouseStuff, error) {
	collection := models.Client.Collection("warehouse_stuffs")
	var stuff models.WarehouseStuff
	var stuffs []models.WarehouseStuff

	cur, err := collection.Find(context.TODO(), bson.M{"warehouse_id": warehouse_id, "com_id": com_id})
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&stuff)
		if err != nil {
			return nil, err
		}
		stuffs = append(stuffs, stuff)
	}
	return stuffs, nil
}

// 查找一批仓库的职员
func FindWarehouseStuffs(warehouse_id []int64, com_id int64) (map[int64][]models.WarehouseStuff, error) {
	if len(warehouse_id) == 0 {
		return nil, nil
	}
	collection := models.Client.Collection("warehouse_stuffs")
	cur, err := collection.Find(context.TODO(), bson.M{"warehouse_id": bson.M{"$in": warehouse_id}, "com_id": com_id})
	// 返回数据格式 map[warehouse_id][]warehouseStuffs
	var stuff models.WarehouseStuff
	stuffs := make(map[int64][]models.WarehouseStuff)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&stuff)
		if err != nil {
			return nil, err
		}
		stuffs[stuff.WarehouseId] = append(stuffs[stuff.WarehouseId], stuff)
	}
	return stuffs, nil
}

// 添加仓库
func AddWarehouse(warehouse models.Warehouse, stuffs []interface{}) (error) {
	wosCollection := models.Client.Collection("warehouse")
	stuffsCollection := models.Client.Collection("warehouse_stuffs")

	_, err := wosCollection.InsertOne(context.TODO(), warehouse)
	if err != nil {
		return err
	}
	if len(stuffs) == 0 {
		return nil
	}
	_, err = stuffsCollection.InsertMany(context.TODO(), stuffs)
	if err != nil {
		wosCollection.DeleteOne(context.TODO(), bson.M{"warehouse_id": warehouse.ID})
		return err
	}
	return nil
}

// 更新仓库信息，所有的仓库信息都会被更新
func UpdateWarehouse(warehouse_id, com_id int64, warehouse interface{}, stuffs []interface{}) (error) {
	wosCollection := models.Client.Collection("warehouse")
	stuffsCollection := models.Client.Collection("warehouse_stuffs")
	_, err := wosCollection.UpdateOne(context.TODO(), bson.M{"warehouse_id": warehouse_id, "com_id": com_id}, bson.M{"$set": warehouse})
	if err != nil {
		return err
	}
	// 把仓库职员表中的数据删除，再填充新数据
	_, err = stuffsCollection.DeleteMany(context.TODO(), bson.M{"warehouse_id": warehouse_id, "com_id": com_id})
	if err != nil {
		return err
	}
	_, err = stuffsCollection.InsertMany(context.TODO(), stuffs)
	if err != nil {
		return err
	}
	return nil
}

// 更新仓库商品记录
func UpdateWarehouseProduct(warehouseId, comId int64, product []int64) error {
	collection := models.Client.Collection("warehouse")
	filter := bson.M{}
	filter["warehouse_id"] = warehouseId
	filter["com_id"] = comId
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"product": product}})
	if err != nil {
		return err
	}
	return nil
}
