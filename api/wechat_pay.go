package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"gitee.com/xiaochengtech/wechat/constant"
	"gitee.com/xiaochengtech/wechat/util"
	"github.com/beevik/etree"
	"net/http"
	"sort"
	"strconv"

	"encoding/json"
	"errors"
)
const (
	// 返回结果
	ResponseSuccess = "SUCCESS" // 成功，通信标识或业务结果
	ResponseFail    = "FAIL"    // 失败，通信标识或业务结果

	// 服务基地址
	BaseUrl        = "https://api.mch.weixin.qq.com/"            // (生产环境) 微信支付的基地址
	BaseUrlSandbox = "https://api.mch.weixin.qq.com/sandboxnew/" // (沙盒环境) 微信支付的基地址

	// 服务模式
	ServiceTypeNormalDomestic      = 1 // 境内普通商户
	ServiceTypeNormalAbroad        = 2 // 境外普通商户
	ServiceTypeFacilitatorDomestic = 3 // 境内服务商
	ServiceTypeFacilitatorAbroad   = 4 // 境外服务商
	ServiceTypeBankServiceProvidor = 5 // 银行服务商

	// 支付类型
	TradeTypeApplet   = "JSAPI"    // 小程序支付
	TradeTypeJsApi    = "JSAPI"    // JSAPI支付
	TradeTypeApp      = "APP"      // APP支付
	TradeTypeH5       = "MWEB"     // H5支付
	TradeTypeNative   = "NATIVE"   // Native支付
	TradeTypeMicropay = "MICROPAY" // 付款码支付

	// 返回消息
	ResponseMessageOk = "OK" // 返回成功信息
)


// 微信支付的整体配置
type WechatConfig struct {
	AppId    string // 微信分配的公众账号ID
	SubAppId string // 微信分配的子商户公众账号ID
	MchId    string // 微信支付分配的商户号
	SubMchId string // 微信支付分配的子商户号，开发者模式下必填
}

// 微信支付客户端配置
type WechatClient struct {
	// 可公开参数
	Config      WechatConfig // 配置信息
	ServiceType int    // 服务模式
	IsProd      bool   // 是否是生产环境
	// 保密参数
	apiKey       string       // API Key
	certFilepath string       // 证书目录
	certClient   *http.Client // 带证书的http连接池
}


// 返回结果的通信标识
type ResponseModel struct {
	ReturnCode string `xml:"return_code"` // SUCCESS/FAIL 此字段是通信标识，非交易标识，交易是否成功需要查看result_code来判断
	ReturnMsg  string `xml:"return_msg"`  // 返回信息，如非空，为错误原因：签名失败/参数格式校验错误
}

// 业务返回结果的错误信息
type ServiceResponseModel struct {
	AppId      string `xml:"appid"`        // 微信分配的公众账号ID
	MchId      string `xml:"mch_id"`       // 微信支付分配的商户号
	SubAppId   string `xml:"sub_appid"`    // (服务商模式) 微信分配的子商户公众账号ID
	SubMchId   string `xml:"sub_mch_id"`   // (服务商模式) 微信支付分配的子商户号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串，不长于32位
	Sign       string `xml:"sign"`         // 签名，详见签名生成算法
	ResultCode string `xml:"result_code"`  // SUCCESS/FAIL
	ErrCode    string `xml:"err_code"`     // 详细参见第6节错误列表
	ErrCodeDes string `xml:"err_code_des"` // 错误返回的信息描述
}

// 场景信息模型
type SceneInfoModel struct {
	ID       string `json:"id"`        // 门店唯一标识
	Name     string `json:"name"`      // 门店名称
	AreaCode string `json:"area_code"` // 门店所在地行政区划码，详细见《最新县及县以上行政区划代码》
	Address  string `json:"address"`   // 门店详细地址
}


// 统一下单的参数
type UnifiedOrderBody struct {
	NonceStr       string `json:"nonce_str"`             // 随机字符串，长度要求在32位以内
	Sign           string `json:"sign"`                  // 通过签名算法计算得出的签名值
	SignType       string `json:"sign_type,omitempty"`   // 签名类型，目前支持HMAC-SHA256和MD5，默认为MD5
	DeviceInfo     string `json:"device_info,omitempty"` // (非必填) 终端设备号(门店号或收银设备ID)，注意：PC网页或JSAPI支付请传"WEB"
	Body           string `json:"body"`                  // 商品描述交易字段格式根据不同的应用场景建议按照以下格式上传： （1）PC网站——传入浏览器打开的网站主页title名-实际商品名称，例如：腾讯充值中心-QQ会员充值；（2） 公众号——传入公众号名称-实际商品名称，例如：腾讯形象店- image-QQ公仔；（3） H5——应用在浏览器网页上的场景，传入浏览器打开的移动网页的主页title名-实际商品名称，例如：腾讯充值中心-QQ会员充值；（4） 线下门店——门店品牌名-城市分店名-实际商品名称，例如： image形象店-深圳腾大- QQ公仔）（5） APP——需传入应用市场上的APP名字-实际商品名称，天天爱消除-游戏充值。
	Detail         string `json:"detail,omitempty"`      // TODO (非必填) 商品详细描述，对于使用单品优惠的商户，该字段必须按照规范上传，详见"单品优惠参数说明"
	Attach         string `json:"attach,omitempty"`      // (非必填) 附加数据，在查询API和支付通知中原样返回，该字段主要用于商户携带订单的自定义数据
	OutTradeNo     string `json:"out_trade_no"`          // 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*且在同一个商户号下唯一。详见商户订单号
	FeeType        string `json:"fee_type,omitempty"`    // (非必填) 符合ISO 4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
	TotalFee       int    `json:"total_fee"`             // 订单总金额，单位为分，只能为整数，详见支付金额
	SpbillCreateIP string `json:"spbill_create_ip"`      // 支持IPV4和IPV6两种格式的IP地址。调用微信支付API的机器IP
	TimeStart      string `json:"time_start,omitempty"`  // (非必填) 订单生成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010。其他详见时间规则
	TimeExpire     string `json:"time_expire,omitempty"` // (非必填) 订单失效时间，格式为yyyyMMddHHmmss，如2009年12月27日9点10分10秒表示为20091227091010。订单失效时间是针对订单号而言的，由于在请求支付的时候有一个必传参数prepay_id只有两小时的有效期，所以在重入时间超过2小时的时候需要重新请求下单接口获取新的prepay_id。其他详见时间规则。建议：最短失效时间间隔大于1分钟
	GoodsTag       string `json:"goods_tag,omitempty"`   // TODO (非必填) 订单优惠标记，代金券或立减优惠功能的参数，说明详见代金券或立减优惠
	NotifyUrl      string `json:"notify_url"`            // 接收微信支付异步通知回调地址，通知url必须为直接可访问的url，不能携带参数。
	TradeType      string `json:"trade_type"`            // JSAPI-JSAPI支付 NATIVE-Native支付 APP-APP支付 说明详见参数规定
	ProductId      string `json:"product_id,omitempty"`  // (非必填) trade_type=NATIVE时，此参数必传。此id为二维码中包含的商品ID，商户自行定义。
	LimitPay       string `json:"limit_pay,omitempty"`   // (非必填) no_credit：指定不能使用信用卡支付
	OpenId         string `json:"openid,omitempty"`      // (非必填) trade_type=JSAPI，此参数必传，用户在主商户appid下的唯一标识。openid和sub_openid可以选传其中之一，如果选择传sub_openid,则必须传sub_appid。下单前需要调用【网页授权获取用户信息】接口获取到用户的Openid。
	SubOpenId      string `json:"sub_openid,omitempty"`  // (非必填) trade_type=JSAPI，此参数必传，用户在子商户appid下的唯一标识。openid和sub_openid可以选传其中之一，如果选择传sub_openid,则必须传sub_appid。下单前需要调用【网页授权获取用户信息】接口获取到用户的Openid。
	Receipt        string `json:"receipt,omitempty"`     // (非必填) Y，传入Y时，支付成功消息和支付详情页将出现开票入口。需要在微信支付商户平台或微信公众平台开通电子发票功能，传此字段才可生效
	SceneInfo      string `json:"scene_info,omitempty"`  // (非必填) 该字段用于上报场景信息，目前支持上报实际门店信息。该字段为JSON对象数据，对象格式为{"store_info":{"id": "门店ID","name": "名称","area_code": "编码","address": "地址" }} ，字段详细说明请点击行前的+展开
	// 用于生成SceneInfo
	SceneInfoModel *SceneInfoModel `json:"-"`
}

// 统一下单的返回值
type UnifiedOrderResponse struct {
	ResponseModel
	// 当return_code为SUCCESS时
	ServiceResponseModel
	DeviceInfo string `xml:"device_info"` // 调用接口提交的终端设备号
	// 当return_code 和result_code都为SUCCESS时
	TradeType string `xml:"trade_type"` // JSAPI-公众号支付 NATIVE-Native支付 APP-APP支付 说明详见参数规定
	PrepayId  string `xml:"prepay_id"`  // 微信生成的预支付回话标识，用于后续接口调用中使用，该值有效期为2小时
	CodeUrl   string `xml:"code_url"`   // trade_type=NATIVE时有返回，此url用于生成支付二维码，然后提供给用户进行扫码支付。注意：code_url的值并非固定，使用时按照URL格式转成二维码即可
	MWebUrl   string `xml:"mweb_url"`   // mweb_url为拉起微信支付收银台的中间页面，可通过访问该url来拉起微信客户端，完成支付，mweb_url的有效期为5分钟。
}

// 支付结果通知的参数
type NotifyPayBody struct {
	ResponseModel
	// 当return_code为SUCCESS时
	ServiceResponseModel
	DeviceInfo         string `xml:"device_info"`          // 微信支付分配的终端设备号
	IsSubscribe        string `xml:"is_subscribe"`         // 用户是否关注公众账号(机构商户不返回)
	SubIsSubscribe     string `xml:"sub_is_subscribe"`     // (服务商模式) 用户是否关注子公众账号(机构商户不返回)
	OpenId             string `xml:"openid"`               // 用户在商户appid下的唯一标识
	SubOpenId          string `xml:"sub_openid"`           // (服务商模式) 用户在子商户appid下的唯一标识
	TradeType          string `xml:"trade_type"`           // 交易类型
	BankType           string `xml:"bank_type"`            // 银行类型，采用字符串类型的银行标识，银行类型见附表
	TotalFee           int    `xml:"total_fee"`            // 订单总金额，单位为分
	FeeType            string `xml:"fee_type"`             // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
	CashFee            int    `xml:"cash_fee"`             // 现金支付金额订单现金支付金额，详见支付金额
	CashFeeType        string `xml:"cash_fee_type"`        // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
	SettlementTotalFee int    `xml:"settlement_total_fee"` // 应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额。
	CouponFee          int    `xml:"coupon_fee"`           // 代金券或立减优惠金额<=订单总金额，订单总金额-代金券或立减优惠金额=现金支付金额，详见支付金额
	CouponCount        int    `xml:"coupon_count"`         // 代金券或立减优惠使用数量
	TransactionId      string `xml:"transaction_id"`       // 微信支付订单号
	OutTradeNo         string `xml:"out_trade_no"`         // 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一。
	Attach             string `xml:"attach"`               // 商家数据包，原样返回
	TimeEnd            string `xml:"time_end"`             // 支付完成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010。其他详见时间规则
	// 使用coupon_count的序号生成的优惠券项
	Coupons []CouponResponseModel `xml:"-"`
}

// 返回结果中的优惠券条目信息
type CouponResponseModel struct {
	CouponId   string // 代金券或立减优惠ID
	CouponType string // CASH-充值代金券 NO_CASH-非充值优惠券 开通免充值券功能，并且订单使用了优惠券后有返回
	CouponFee  int64  // 单个代金券或立减优惠支付金额
}

// 微信通知的结果返回值
type NotifyResponseModel struct {
	ReturnCode string // SUCCESS/FAIL
	ReturnMsg  string // 返回信息，如非空，为错误原因，或OK
}


// 初始化微信支付客户端
func NewWechatClient(isProd bool, serviceType int, apiKey string, certFilepath string, config WechatConfig) (client *WechatClient) {
	client = new(WechatClient)
	client.Config = config
	client.ServiceType = serviceType
	client.IsProd = isProd
	client.apiKey = apiKey
	client.certFilepath = certFilepath
	return client
}



// 统一下单
func (c *WechatClient) UnifiedOrder(body UnifiedOrderBody) (wxRsp UnifiedOrderResponse, err error) {
	// 处理参数
	if body.SceneInfoModel != nil {
		body.SceneInfo = util.MarshalJson(*body.SceneInfoModel)
	}
	// 业务逻辑
	bytes, err := c.doWeChat("pay/unifiedorder", body)
	if err != nil {
		return
	}
	// 结果校验
	if err = c.doVerifySign(bytes, true); err != nil {
		return
	}
	// 解析返回值
	err = xml.Unmarshal(bytes, &wxRsp)
	return
}

// 向微信发送请求
func (c *WechatClient) doWeChat(relativeUrl string, bodyObj interface{}) (bytes []byte, err error) {
	// 转换参数
	body, err := c.buildBody(bodyObj)
	if err != nil {
		return
	}
	// 发起请求
	bytes, err = util.HttpPostXml(c.URL(relativeUrl), util.GenerateXml(body))
	return
}

// 验证微信返回的结果签名
func (c *WechatClient) doVerifySign(xmlStr []byte, breakWhenFail bool) (err error) {
	// 生成XML文档
	doc := etree.NewDocument()
	if err = doc.ReadFromBytes(xmlStr); err != nil {
		return
	}
	root := doc.SelectElement("xml")
	// 验证return_code
	retCode := root.SelectElement("return_code").Text()
	if retCode != ResponseSuccess && breakWhenFail {
		return
	}
	// 遍历所有Tag，生成Map和Sign
	result, targetSign := make(map[string]interface{}), ""
	for _, elem := range root.ChildElements() {
		// 跳过空值
		if elem.Text() == "" {
			continue
		}
		if elem.Tag != "sign" {
			result[elem.Tag] = elem.Text()
		} else {
			targetSign = elem.Text()
		}
	}
	// 获取签名类型
	signType := constant.SignTypeMD5
	if result["sign_type"] != nil {
		signType = result["sign_type"].(string)
	}
	// 生成签名
	var sign string
	if c.IsProd {
		sign = c.localSign(result, signType, c.apiKey)
	}
	// 验证
	if targetSign != sign {
		err = errors.New("签名无效")
	}
	return
}

// 构建Body
func (c *WechatClient) buildBody(bodyObj interface{}) (body map[string]interface{}, err error) {
	// 将bodyObj转换为map[string]interface{}类型
	bodyJson, _ := json.Marshal(bodyObj)
	body = make(map[string]interface{})
	_ = json.Unmarshal(bodyJson, &body)
	// 添加固定参数
	body["appid"] = c.Config.AppId
	body["mch_id"] = c.Config.MchId
	if c.IsFacilitator() {
		body["sub_appid"] = c.Config.SubAppId
		body["sub_mch_id"] = c.Config.SubMchId
	}
	nonceStr := util.RandomString(32)
	body["nonce_str"] = nonceStr
	// 生成签名
	signType, _ := body["sign_type"].(string)
	var sign string
	if c.IsProd {
		sign = c.localSign(body, signType, c.apiKey)
	}
	body["sign"] = sign
	fmt.Println("Sign with ", body["sign"])
	return
}

// 拼接完整的URL
func (c *WechatClient) URL(relativePath string) string {
	if c.IsProd {
		return BaseUrl + relativePath
	} else {
		return BaseUrlSandbox + relativePath
	}
}

// 本地通过支付参数计算签名值
func (c *WechatClient) localSign(body map[string]interface{}, signType string, apiKey string) string {
	signStr := c.sortSignParams(body, apiKey)
	fmt.Println("signType: ", signType)
	fmt.Println("signStr: ", signStr)
	return util.SignWithType(signType, signStr, apiKey)
}

// 获取根据Key排序后的请求参数字符串
func (c *WechatClient) sortSignParams(body map[string]interface{}, apiKey string) string {
	keyList := make([]string, 0)
	for k := range body {
		if body[k] == "" {
			continue
		}
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)
	buffer := new(bytes.Buffer)
	for _, k := range keyList {
		s := fmt.Sprintf("%s=%s&", k, fmt.Sprintf("%v", body[k]))
		buffer.WriteString(s)
	}
	buffer.WriteString(fmt.Sprintf("key=%s", apiKey))
	return buffer.String()
}


// 是否是服务商模式
func (c *WechatClient) IsFacilitator() bool {
	switch c.ServiceType {
	case ServiceTypeFacilitatorDomestic, ServiceTypeFacilitatorAbroad, ServiceTypeBankServiceProvidor:
		return true
	default:
		return false
	}
}

// 小程序支付，统一下单获取支付参数后，再次计算出小程序用的paySign
func GetAppletPaySign(
	appId string,
	nonceStr string,
	prepayId string,
	signType string,
	timeStamp string,
	apiKey string,
) (paySign string) {
	// 原始字符串
	signStr := fmt.Sprintf("appId=%s&nonceStr=%s&package=%s&signType=%s&timeStamp=%s&key=%s",
		appId, nonceStr, prepayId, signType, timeStamp, apiKey)
	fmt.Println("applet pay sign string: ", signStr)
	// 加密签名
	paySign = util.SignWithType(signType, signStr, apiKey)
	fmt.Println("applet pay sign: ", paySign)
	return
}

type NotifyPayHandler func(NotifyPayBody) error

// 支付结果通知
func (c *WechatClient) NotifyPay(handler NotifyPayHandler, requestBody []byte) (rspBody string, err error) {
	// 验证Sign
	if err = c.doVerifySign(requestBody, false); err != nil {
		return
	}
	// 解析参数
	var body NotifyPayBody
	if err = c.payNotifyParseParams(requestBody, &body); err != nil {
		return
	}
	// 调用外部处理
	if err = handler(body); err != nil {
		return
	}
	// 返回处理结果
	rspModel := NotifyResponseModel{
		ReturnCode: ResponseSuccess,
		ReturnMsg:  ResponseMessageOk,
	}
	rspBody = rspModel.ToXmlString()
	return
}

// 支付结果通知-解析XML参数
func (c *WechatClient) payNotifyParseParams(xmlStr []byte, body *NotifyPayBody) (err error) {
	if err = xml.Unmarshal(xmlStr, &body); err != nil {
		return
	}
	// 解析CouponCount的对应项
	if body.CouponCount > 0 {
		doc := etree.NewDocument()
		if err = doc.ReadFromBytes(xmlStr); err != nil {
			return
		}
		root := doc.SelectElement("xml")
		for i := 0; i < body.CouponCount; i++ {
			m := NewCouponResponseModel(root, "coupon_id_%d", "coupon_type_%d", "coupon_fee_%d", i)
			body.Coupons = append(body.Coupons, m)
		}
	}
	return
}

func (m *NotifyResponseModel) ToXmlString() string {
	buffer := new(bytes.Buffer)
	buffer.WriteString("<xml>")
	buffer.WriteString(fmt.Sprintf("<return_code><![CDATA[%s]]></return_code>", m.ReturnCode))
	buffer.WriteString(fmt.Sprintf("<return_msg><![CDATA[%s]]></return_msg>", m.ReturnMsg))
	buffer.WriteString("</xml>")
	return buffer.String()
}

// 在XML节点树中，查找labels对应的
func NewCouponResponseModel(
	doc *etree.Element,
	idFormat string,
	typeFormat string,
	feeFormat string,
	numbers ...interface{},
) (m CouponResponseModel) {
	idName := fmt.Sprintf(idFormat, numbers...)
	typeName := fmt.Sprintf(typeFormat, numbers...)
	feeName := fmt.Sprintf(feeFormat, numbers...)
	m.CouponId = doc.SelectElement(idName).Text()
	m.CouponType = doc.SelectElement(typeName).Text()
	m.CouponFee, _ = strconv.ParseInt(doc.SelectElement(feeName).Text(), 10, 64)
	return
}
