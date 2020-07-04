package api

import (
	"github.com/gin-gonic/gin"
	"jxc/serializer"
	"net/http"
)

type AuthData struct {
	Username string	`valid:"Required; MaxSize(50)"`
	Password string `valid:"Required; MaxSize(50)"`
}

func GetAuth(c *gin.Context) {
/*	username := c.Query("username")
	password := c.Query("password")

	//valid := validator.New()
	//a := AuthData{Username:username, Password: password}
	ok, _ := true, true

	data := make(map[string]interface{})
	code := -1 // invalid params code

	if ok {
		isExist := models.CheckAuth(username, password)
		if isExist {
			token, err := auth.GenerateToken(username, 1)
			if err != nil {
				code = -1 // error auth token
			} else {
				data["token"] = token
				code = 200
			}
		} else {
			code = -1 // error auth
		}
	} else {
		//for _, err := valid.Errors {
		//	log.Println(err.Key, err.Message)
		//}
	}*/

	c.JSON(http.StatusOK, serializer.Response{
		Code: -1,
		Msg: "Auth failed",
	})
}

//type Transferer interface {
//	TransferMoney(id int, buyerID int, sellerID int, amount float64) error
//}
//
//type transaction struct {
//	ID       string
//	BuyerID  int
//	SellerID int
//	Amount   float64
//	createdAt time.Time
//	Status TransactionStatus
//	// 增加了一个存放接口的属性
//	transferer Transferer
//}
//
//func New(buyerID, sellerID int, amount float64, transferer Transferer) *transaction {
//	return &transaction{
//		ID:         IdGenerator.generate(),
//		BuyerID:    buyerID,
//		SellerID:   sellerID,
//		Amount:     amount,
//		createdAt:  time.Now(),
//		Status:     TO_BE_EXECUTD,
//		transferer: transferer, // 注入进 transaction 类中
//	}
//}
//
//func (t *transaction) Execute() bool {
//	if t.Status == Executed {
//		return true
//	}
//	if time.Now() - t.createdAt > 24.hours { // 交易有有效期
//		t.Status = Expired
//		return false
//	}
//	t.transferer.TransferMoney(id, t.BuyerID, t.SellerID, t.Amount)
//	t.Status = Executed
//	return true
//}
//
//type MockedClient struct {
//	responseError error // 实例化的时候可以将期望的返回值保存进来
//}
//
//func (m *MockedClient) TransferMoney(id int, buyerID int, sellerID int, amount float64) error {
//	return m.responseError
//}
//
//func Test_transaction_Execute(t *testing.T) {
//	// 实例化一个可以自由控制结果的 client
//	transferer := &MockedClient{
//		responseError: errors.New("insufficient balance"),
//	}
//	tnx := New(buyerID, sellerID, amount, transferer)
//	if succeeded := tnx.Execute(); succeeded != false {
//		t.Errorf("Execute() = %v, want %v", succeeded, false)
//	}
//}