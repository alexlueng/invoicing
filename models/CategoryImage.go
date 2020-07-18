package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 商品图片存储表
type CategoryImage struct {
	ImageID    int64  `json:"image_id" bson:"image_id"`
	ComID      int64  `json:"com_id" bson:"com_id"`
	CategoryID int64  `json:"category_id" bson:"category_id"`
	LocalPath  string `json:"local_path" bson:"local_path"` // 本地存放路径
	CloudPath  string `json:"cloud_path" bson:"cloud_path"` // 云上存储路径
	IsDelete   bool   `json:"is_delete" bson:"is_delete"`
	Order      int64  `json:"order" bson:"order"`
}

func getCategoryImageCollection() *mongo.Collection {
	return Client.Collection("category_image")
}

func (c *CategoryImage) Add() error {
	_, err := getCategoryCollection().InsertOne(context.TODO(), c)
	return err
}

type CategoryImageResult struct {
	CategoryImage []CategoryImage `json:"category_image"`
}

func SelectCategoryImageByComID(comID int64) (*CategoryImageResult, error)  {

	cur, err := getCategoryImageCollection().Find(context.TODO(), bson.D{{"com_id", comID}})
	if err != nil {
		return nil , err
	}

	var res = new(CategoryImageResult)
	for cur.Next(context.TODO()) {
		var c CategoryImage
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		res.CategoryImage = append(res.CategoryImage, c)
	}
	return res, nil
}





