package wxapp

import (
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

type DetailService struct {
	ProductID int64 `json:"product_id"`
}

// 商品详情列表
func ProductDetail(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var detailSrv DetailService
	if err := c.ShouldBindJSON(&detailSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	product, err := models.GetProductByID(claims.ComId, detailSrv.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}

	detail, err := models.GetProductDetailByID(claims.ComId, detailSrv.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product detail",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeError,
		Msg:  "Product detail",
		Data: map[string]interface{}{
			"product": product,
			"detail":  detail,
		},
	})
}
