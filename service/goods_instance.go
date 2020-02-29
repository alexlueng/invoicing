package service

import (
	"context"
	"jxc/models"
)

// 商品实例

// 添加商品实例
func AddGoodsInstance(wosInstance []interface{}) (error) {
	collection := models.Client.Collection("goods_instance")
	_, err := collection.InsertMany(context.TODO(), wosInstance)
	if err != nil {
		return nil
	}
	return nil
}
