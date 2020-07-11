package wxapp

type Offical struct {
	ComID       int64  `json:"com_id" bson:"com_id"`
	OriginID    string `json:"origin_id" bson:"origin_id"`       // 原始ID
	WechatName  string `json:"wechat_name" bson:"wechat_name"`   // 微信号
	OfficalType int64  `json:"offical_type" bson:"offical_type"` // 公众号类型
	QRCode      string `json:"qr_code" bson:"qr_code"`           // 二维码地址
	Domain      string `json:"domain" bson:"domain"`             // 接口域名
	AppID       string `json:"app_id" bson:"app_id"`
	AppSecret   string `json:"app_secret" bson:"app_secret"`
}
