package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	ComID        int64    `json:"com_id" bson:"com_id"`
	ProductID    int64    `json:"product_id" bson:"product_id"`
	Product      string   `json:"product" bson:"product" form:"product"`
	Units        string   `json:"units" bson:"units"`
	URLS         []string `json:"urls" bson:"urls"`
	LocalPaths   []string `json:"local_paths" bson:"local_paths"` // 图片本地存放的路径
	Num          int64    `json:"num" bson:"num"`     // 销量
	Stock        int64    `json:"stock" bson:"stock"` // 库存
	DefaultPrice float64  `json:"default_price" bson:"default_price" form:"default_price"`
	CusPrice     []int64  `json:"cus_price" bson:"cus_price"`
	SupPrice     []int64  `json:"sup_price" bson:"sup_price"`
	CatID        int64    `json:"cat_id" bson:"cat_id"`       // 商品分类
	MinAlert     int64    `json:"min_alert" bson:"min_alert"` // 最低库存预警
	Discribe     string   `json:"discribe" bson:"discribe"`   // 商品描述
	Preferred    bool     `json:"preferred" bson:"preferred"` // 优选商品
	Recommand    bool     `json:"recommand" bson:"recommand"` // 推荐商品
	Tags         []string `json:"tags" bson:"tags"`           // 商品标签
}

type PriceOfSupplier struct {
	SupplierID int64   `json:"supplier_id"`
	Supplier   string  `json:"supplier"`
	Price      float64 `json:"price"`
}

type ProductSupplier struct {
	ID   int    `json:"supplier_id" bson:"supplier_id"`
	Name string `json:"supplier_name" bson:"supplier_name"`
}

type ProductReq struct {
	BaseReq
	Product          string            `json:"product" form:"product"`
	Units            string            `json:"units" form:"units"`
	PriceOfSuppliers []PriceOfSupplier `json:"price_of_supplier"`
	InStock          int64             `json:"in_stock"`
}

type ResponseProductData struct {
	Products      []Product   `json:"products"`
	ProductImages interface{} `json:"product_images"`
	Total         int         `json:"total"`
	Pages         int         `json:"pages"`
	Size          int         `json:"size"`
	CurrentPage   int         `json:"current_page"`
}

type ProductList struct {
	ID           int64   `bson:"product_id"`
	Product      string  `bson:"product"`
	CusPrice     []int64 `bson:"cus_price"`
	DefaultPrice float64 `bson:"default_price"`
}

func getProductCollection() *mongo.Collection {
	return Client.Collection("product")
}

func GetProductByID(com_id, product_id int64) (product *Product, err error) {

	filter := bson.M{}
	filter["com_id"] = com_id
	filter["product_id"] = product_id

	err = getProductCollection().FindOne(context.TODO(), filter).Decode(&product)
	if err != nil {
		return nil, err
	}
	return
}

func UpdateQuantityByComIDAndProductID(comID, productID, amount int64) error {
	_, err := getProductCollection().UpdateOne(context.TODO(),
		bson.D{{"com_id", comID}, {"product_id", productID}},
		bson.M{"$inc": bson.M{"num": amount}})
	return err
}

func UpdateCusPriceByProductID(productID, customerID int64) error {
	insertProduct := bson.M{"product_id": productID}
	pushToArray := bson.M{"$addToSet": bson.M{"cus_price": customerID}}
	_, err := getProductCollection().UpdateOne(context.TODO(), insertProduct, pushToArray)
	return err
}

func UpdateProductDefaultPriceByProductID(productID int64, price float64) error {
	filter := bson.M{}
	filter["product_id"] = productID
	_, err := getProductCollection().UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"default_price": price}})
	return err
}

func SelectProductListByComID(comID int64, req CustomerProductPriceReq) ([]ProductList, error) {

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
	option.Projection = bson.M{"product_id": 1, "product": 1, "cus_price": 1, "default_price": 1, "_id": 0}

	filter := bson.M{}
	filter["com_id"] = comID

	var allProducts []ProductList
	cur, err := getProductCollection().Find(context.TODO(), filter, option)

	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var result ProductList
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		allProducts = append(allProducts, result)
	}
	return allProducts, nil
}

