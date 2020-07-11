package wxapp

// 页面广告表

type Advertisement struct {
	ComID    int64  `json:"com_id" bson:"com_id"`
	AdID     int64  `json:"ad_id" bson:"ad_id"`
	Title    string `json:"title" bson:"title"`       // 标题
	Content  string `json:"content" bson:"content"`   // 内容
	IsUsing  bool   `json:"is_using" bson:"is_using"` // 是否启用
	Location string `json:"location" bson:"location"` // 展示位置
	Link     string `json:"link" bson:"link"`         // 链接
}
