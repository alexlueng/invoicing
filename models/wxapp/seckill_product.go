package wxapp

// 秒杀商品表

type SeckillProduct struct {
	ComID         int64   `json:"com_id"`
	ID            int64   `json:"id" bson:"id"`
	Product       string  `json:"product" bson:"product"`
	Price         float64 `json:"price" bson:"price"`                   // 原价
	DiscountPrice float64 `json:"discount_price" bson:"discount_price"` // 折扣价
	ImageURL      string  `json:"image_url" bson:"image_url"`           // 商品图
	Comment       string  `json:"comment" bson:"comment"`               // 标注
	IsShow        bool    `json:"is_show" bson:"is_show"`               // 是否展示
}
