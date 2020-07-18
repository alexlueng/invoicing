package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"jxc/models"
	"jxc/serializer"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	payClient = NewPayClient(nil)

	ErrCode_Success = int64(0)
	ErrCode_ParseQuery = int64(1)
	ErrCode_Param = int64(2)
	ErrCode_PayData = int64(3)

	ErrCode_Product = int64(101)
)

//func (srv *DefaultPayService) Start()  {
//
//	// login
//	http.HandleFunc("/Login", srv.handlerLogin)
//
//}

func handlerError(w http.ResponseWriter, err int64, msg string) func() {

	data := models.NewPayData()
	return func() {

		data.Set("errcode", err)
		data.Set("errmsg", msg)
		io.WriteString(w, data.ToJson())

		log.Println(fmt.Sprintf("errcode: %d, errmsg: %s", err, msg))
	}
}

func handlerErrorXML(w http.ResponseWriter, err string, msg string) func() {

	data := models.NewPayData()
	return func() {

		data.Set("return_code", err)
		data.Set("return_msg", msg)
		io.WriteString(w, string(data.ToXml()))

		log.Println(fmt.Sprintf("return_code: %s, return_msg: %s", err, msg))
	}
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {

	reqData, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		handlerError(w, ErrCode_ParseQuery, err.Error())()
		return
	}

	jscode := reqData.Get("code")
	if jscode == "" {
		handlerError(w, ErrCode_Param, "error param")()
		return
	}

	rspData, err := payClient.WechatLogin(jscode)
	if err != nil {
		handlerError(w, ErrCode_PayData, err.Error())()
		return
	}

	log.Println("-----------------handlerLogin succeed-------------------")
	log.Println("openid: "+rspData.Get("openid"))
	log.Println("session_key: "+rspData.Get("session_key"))

	io.WriteString(w, rspData.ToJson())

	//payListener.HandleLogicLogin(rspData)
}

// 统一下单接口
func UnifiedOrder(c *gin.Context) {

	reqData, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	openId := reqData.Get("openid")
	productId := reqData.Get("productid")
	billIp := reqData.Get("ip")
	if openId == "" || productId == "" || billIp == "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	//product := payListener.HandleLogicProduct(productId)
	//if product == nil {
	//	srv.handlerError(w, ErrCode_Product, "no goods info")()
	//	return
	//}

	tradeNo := NonceStr()

	apiData := models.NewPayData()
	apiData.Set("openid", openId)
	apiData.Set("product_id", productId)
	//apiData.Set("body", product.Get("body"))
	//apiData.Set("total_fee", product.Get("total_fee"))
	//apiData.Set("detail", product.Get("detail"))
	//apiData.Set("fee_type", product.Get("fee_type"))
	apiData.Set("trade_type", TradeType_JsApi)
	apiData.Set("out_trade_no", tradeNo)
	apiData.Set("spbill_create_ip", billIp)
	apiData.Set("notify_url", models.PayConfigInstance.NotifyUrl())
	apiData.Set("time_start", FormatTime(time.Now()))
	apiData.Set("time_expire", FormatTime(time.Now().Add(time.Minute * 10)))
	apiData.Set("nonce_str", NonceStr())

	rapiData, err := ApiUnifiedOrder(payClient, apiData)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	timeStamp := TimeStamp()
	nonceStr := NonceStr()
	packageStr := "prepay_id=" + rapiData.Get("prepay_id")

	rspData := models.NewPayData()
	rspData.Set("appId", models.PayConfigInstance.AppId())
	rspData.Set("timeStamp", timeStamp)
	rspData.Set("nonceStr", nonceStr)
	rspData.Set("package", packageStr)
	rspData.Set("signType", SignType_MD5)

	paySign := rspData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_MD5)

	rspData.Set("errcode", ErrCode_Success)
	rspData.Set("errmsg", RCSuccess)
	rspData.Set("paySign", paySign)
	rspData.Set("tradeNo", tradeNo)

	io.WriteString(c.Writer, rspData.ToJson())

	log.Println("-----------------handlerUnifiedOrder succeed-------------------")
}

// 支付结果
func PayResult(c *gin.Context) {

	log.Println("-----------------handlerPayResult succeed-------------------")

	resultData := models.NewPayData()
	err := resultData.FromXml(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	defer c.Request.Body.Close()

	if !resultData.IsSet("transaction_id") ||
		resultData.Get("transaction_id") == "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	apiData := models.NewPayData()
	apiData.Set("transaction_id", resultData.Get("transaction_id"))
	rapiData, err := ApiOrderQuery(payClient, apiData)
	if err != nil {
		//handlerErrorXML(w, RCFail, "error order query")()
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	} else {
		if rapiData.Get("return_code") == RCSuccess &&
			rapiData.Get("result_code") == RCSuccess {
			//handlerErrorXML(w, RCSuccess, "OK")()
				c.JSON(http.StatusOK, serializer.Response{
					Code: 200,
					Msg:  "OK",
				})
				return
			// TODO:
			//payListener.HandleLogicPay()
		} else {
			//handlerErrorXML(w, RCFail, "error order query")()
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "error order query",
			})
		}
	}
}

func OrderQuery(c *gin.Context) {

	reqData, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	tradeNo := reqData.Get("tradeno")
	if tradeNo == "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "trade no error",
		})
	}

	apiData := models.NewPayData()
	apiData.Set("out_trade_no", tradeNo)

	rapiData, err := ApiOrderQuery(payClient, apiData)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	rspData := models.NewPayData()
	if rapiData.Get("return_code") == RCSuccess &&
		rapiData.Get("result_code") == RCSuccess {
		rspData.Set("errcode", ErrCode_Success)
	} else {
		rspData.Set("errcode", ErrCode_PayData)
	}
	rspData.Set("errmsg", rapiData.Get("return_msg"))

	io.WriteString(c.Writer, rspData.ToJson())

	log.Println("-----------------handlerOrderQuery succeed-------------------")
}

/*func (srv *DefaultPayService) handlerMicropay(w http.ResponseWriter, r *http.Request) {

	// TODO:
}

func (srv *DefaultPayService) handlerCloseOrder(w http.ResponseWriter, r *http.Request) {

	// TODO:
}

func (srv *DefaultPayService) handlerRefund(w http.ResponseWriter, r *http.Request) {

	// TODO:
}

func (srv *DefaultPayService) handlerReverse(w http.ResponseWriter, r *http.Request) {

	// TODO:
}

func (srv *DefaultPayService) handlerRefundQuery(w http.ResponseWriter, r *http.Request) {

	// TODO:
}

func (srv *DefaultPayService) handlerDownloadBill(w http.ResponseWriter, r *http.Request) {

	// TODO:
}*/




const (
	wxBase_url 	= "https://api.mch.weixin.qq.com/"

	//正式
	wxURL_UnifiedOrder      = wxBase_url + "pay/unifiedorder"                //统一下单
	wxURL_OrderQuery        = wxBase_url + "pay/orderquery"                  //查询订单
	wxURL_Micropay          = wxBase_url + "pay/micropay"                    //提交付款码支付
	wxURL_CloseOrder        = wxBase_url + "pay/closeorder"                  //关闭订单
	wxURL_Refund            = wxBase_url + "secapi/pay/refund"               //申请退款
	wxURL_Reverse           = wxBase_url + "secapi/pay/reverse"              //撤销订单
	wxURL_RefundQuery       = wxBase_url + "pay/refundquery"                 //查询退款
	wxURL_DownloadBill      = wxBase_url + "pay/downloadbill"                //下载对账单

	//支付类型
	TradeType_JsApi  = "JSAPI"
	TradeType_App    = "APP"
	TradeType_H5     = "MWEB"
	TradeType_Native = "NATIVE"

	SignType_MD5         = "MD5"
	SignType_SHA1        = "SHA1"
	SignType_HMAC_SHA256 = "HMAC-SHA256"

	RCSuccess = "SUCCESS"
	RCFail    = "FAIL"
)

func ApiUnifiedOrder(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") ||
		!payData.IsSet("body") ||
		!payData.IsSet("total_fee") ||
		!payData.IsSet("trade_type") ||
		!payData.IsSet("notify_url") ||
		!payData.IsSet("spbill_create_ip"){
		return nil, fmt.Errorf("need pay param")
	}

	if payData.Get("trade_type") == TradeType_JsApi && !payData.IsSet("openid") {
		return nil, fmt.Errorf("need openid")
	}

	if payData.Get("trade_type") == TradeType_Native && !payData.IsSet("product_id") {
		return  nil, fmt.Errorf("need product_id")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("sign_type", SignType_MD5)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_MD5))

	rpayData, err := payClient.PostXML(wxURL_UnifiedOrder, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiOrderQuery(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") &&
		!payData.IsSet("transaction_id") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_OrderQuery, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiMicropay(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") ||
		!payData.IsSet("body") ||
		!payData.IsSet("total_fee") ||
		!payData.IsSet("auth_code") ||
		!payData.IsSet("spbill_create_ip"){
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_Micropay, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiCloseOrder(payClient *PayClient, payData *models.PayData) (*models.PayData, error) {

	if !payData.IsSet("out_trade_no") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_CloseOrder, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiRefund(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") ||
		!payData.IsSet("out_refund_no") ||
		!payData.IsSet("total_fee") ||
		!payData.IsSet("refund_fee") ||
		!payData.IsSet("op_user_id") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_Refund, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiReverse(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_Reverse, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiRefundQuery(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("out_trade_no") &&
		!payData.IsSet("out_refund_no") &&
		!payData.IsSet("transaction_id") &&
		!payData.IsSet("refund_id") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_RefundQuery, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

func ApiDownloadBill(payClient *PayClient, payData *models.PayData) (*models.PayData, error)  {

	if !payData.IsSet("bill_date") {
		return nil, fmt.Errorf("need pay param")
	}

	payData.Set("appid", models.PayConfigInstance.AppId())
	payData.Set("mch_id", models.PayConfigInstance.MchId())
	payData.Set("nonce_str", NonceStr())
	payData.Set("sign_type", SignType_HMAC_SHA256)
	payData.Set("sign", payData.MakeSign(models.PayConfigInstance.ApiKey(), SignType_HMAC_SHA256))

	rpayData, err := payClient.PostXML(wxURL_DownloadBill, payData)
	if err != nil {
		return nil, err
	}

	return rpayData, nil
}

