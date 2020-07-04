package models

// 客户等级表
type Level struct {
	ComID     int64   `json:"com_id" bson:"com_id"`
	LevelID   int64   `json:"level_id" bson:"level_id"`
	LevelName string  `json:"level_name" bson:"level_name"`
	Discount  float64 `json:"discount" bson:"discount"`
}
