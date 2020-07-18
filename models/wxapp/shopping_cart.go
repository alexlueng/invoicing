package wxapp

// 两个表 一个主表 一个详情表

type UserCart struct {
	ComID      int64  `json:"com_id" bson:"com_id"`
	CustomerID int64  `json:"customer_id" bson:"customer_id"`
	CartID     int64  `json:"cart_id" bson:"cart_id"`
	OpenID     string `json:"open_id" bson:"open_id"`
}

/*
	需要考虑的问题：商品删除或下架了，购物车内的商品是否会删除，或是给个标志位，
	让它显示不可用
	购物车中的商品只能一个一个加
*/
type CartItem struct {
	ComID       int64   `json:"com_id" bson:"com_id"`
	CartID      int64   `json:"cart_id" bson:"cart_id"`
	ProductID   int64   `json:"product_id" bson:"product_id"`
	OpenID      string  `json:"open_id" bson:"open_id"`
	ItemID      int64   `json:"item_id" bson:"item_id"`
	ProductName string  `json:"product_name" bson:"product_name"`
	Thumbnail   string  `json:"thumbnail" bson:"thumbnail"` // 商品缩略图
	Num         int64   `json:"num" bson:"num"`
	Price       float64 `json:"price" bson:"price"`
	CreateAt    int64   `json:"create_at" bson:"create_at"`
	IsDelete    bool    `json:"is_delete" bson:"is_delete"`       // 是否在购物车中删除
	IsAvaliable bool    `json:"is_avaliable" bson:"is_avaliable"` // 是否可用
}
