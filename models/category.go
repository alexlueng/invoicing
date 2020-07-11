package models

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"jxc/util"
)

// 商品分类
type Category struct {
	ID           int64  `json:"id" bson:"id"` // 分类ID
	ComID        int64  `json:"com_id" bson:"com_id"`
	CategoryName string `json:"category_name" bson:"category_name"`   // 分类名字
	ParentID     int64  `json:"parent_id" bson:"parent_id"`           // 父级id
	ParentIDPath string `json:"parent_id_path" bson:"parent_id_path"` // 父级id路径
	Level        int64  `json:"level" bson:"level"`                   // 几级分类
	IsDelete     bool   `json:"is_delete" bson:"is_delete"`           // 是否删除
	Comment      string `json:"comment" bson:"comment"`               // 备注
	//Thumbnail    Image  `json:"thumbnail" bson:"thumbnail"`           // 缩略图
}

type CategorySearchResult struct {
	Category []Category `json:"category"`
}

func getCategoryCollection() *mongo.Collection {
	return Client.Collection("category")
}

func (c *Category) Add() error {
	_, err := getCategoryCollection().InsertOne(context.TODO(), c)
	if nil != err {
		util.Log().Error("add category failed, err:", err)
		return err
	}
	return nil
}

func SelectCategoryById(id int64) (*Category, error) {
	var c Category
	err := getCategoryCollection().FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&c)
	if nil != err {
		util.Log().Error("SelectCategoryById failed, err:", err)
		return nil, err
	}
	return &c, nil
}

//查找顶级分类
func SelectTopCategoryByComId(comId int64) (*CategorySearchResult, error) {
	return SelectCategoryByComIdAndParentId(0, comId)
}

//按分类与公司查找
func SelectCategoryByComIdAndParentId(parentId, comId int64) (*CategorySearchResult, error) {
	cur, err := getCategoryCollection().Find(context.TODO(), bson.D{{"parent_id", parentId}, {"com_id", comId}})
	if err != nil {
		return nil, err
	}
	var sr = new(CategorySearchResult)
	for cur.Next(context.TODO()) {
		var category Category
		err := cur.Decode(&category)
		if err != nil {
			return nil, err
		}
		sr.Category = append(sr.Category, category)
	}
	if nil == sr.Category || len(sr.Category) < 1 {
		return nil, errors.New("未获取到分类数据")
	}
	return sr, nil
}

//查找顶级分类下的子类
func SelectCategoryByComIdAndTopCategoryID(comId int64, toplist []int64) (*CategorySearchResult, error) {
	filter := bson.M{}
	if len(toplist) > 0 {
		filter["parent_id"] = bson.M{"$in": toplist}
	}
	filter["com_id"] = comId
	cur, err := getCategoryCollection().Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var sr = new(CategorySearchResult)
	for cur.Next(context.TODO()) {
		var category Category
		err := cur.Decode(&category)
		if err != nil {
			return nil, err
		}
		sr.Category = append(sr.Category, category)
	}
	if nil == sr.Category || len(sr.Category) < 1 {
		return nil, errors.New("未获取到分类数据")
	}
	return sr, nil
}

//按ID与ComID删除分类
func DeleteCategoryByComIdAndId(comID, ID int64) error {
	_, err := getCategoryCollection().DeleteOne(context.TODO(), bson.D{{"id", ID}, {"com_id", comID}})
	if err != nil {
		return err
	}
	return nil
}

//更新分类
func UpdateCategory(comId int64, categoryID int64, categoryName string) error {
	_, err := getCategoryCollection().UpdateOne(context.TODO(), bson.D{{"id", categoryID}, {"com_id", comId}}, bson.M{"$set": bson.M{"category_name": categoryName}})
	return err
}
