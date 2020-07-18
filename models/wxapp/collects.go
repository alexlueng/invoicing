package wxapp

// 用户收藏列表
type UserCollection struct {
	ComID      int64 `json:"com_id" bson:"com_id"`
	CustomerID int64 `json:"customer_id" bson:"customer_id"`
	ProductID  int64 `json:"product_id" bson:"product_id"`
	CollectID  int64 `json:"collect_id" bson:"collect_id"`
	Status     int64 `json:"status" bson:"status"` // 收藏商品状态 1 有效 0 无效
}
