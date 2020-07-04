package models

// 配送方式数据格式
type Delivery struct {
	DeliveryId     int64  `json:"delivery_id" bson:"delivery_id"`        //配送方式id 这个id由前端生成
	ComId          int64  `json:"com_id" bson:"comid"`                   //公司id
	DeliveryCom    string `json:"delivery_com" bson:"deliverycom"`       // 配送公司
	DeliveryPerson string `json:"delivery_person" bson:"deliveryperson"` // 配送员
	Phone          string `json:"phone" bson:"phone"`                    // 配送员电话
	Config         string `json:"config" bson:"config"`                  // 配置参数
	IsUsing        bool   `json:"is_using" bson:"is_using"`              // 是否启用
}

