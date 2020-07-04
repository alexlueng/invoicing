package models

// 商品分类
type Category struct {
	ID           int64  `json:"id" bson:"id"` // 分类ID
	ComID        int64  `json:"com_id" bson:"com_id"`
	CategoryName string `json:"category_name" bson:"category_name"`   // 分类名字
	ParentID     int64  `json:"parent_id" bson:"parent_id"`           // 父级id
	ParentIDPath string `json:"parent_id_path" bson:"parent_id_path"` // 父级id路径
	Level        int64  `json:"level" bson:"level"`                   // 几级分类
	IsDelete     bool   `json:"is_delete" bson:"is_delete"`           // 是否删除
}
