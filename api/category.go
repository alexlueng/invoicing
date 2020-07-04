package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strconv"
)

//商品分类最多分为三级
//添加或者修改分类时, 应注意选择对应的上级

// 添加分类

type AddCateoryService struct {
	Name     string `json:"category_name"`
	IsTop    bool   `json:"is_top"`    // 是否顶级
	ParentID int64  `json:"parent_id"` // 父级ID
}


// 分配一个分类id，查看是否顶级分类，如果是，则parent_id = 0, parent_id_path = 0_id
// 如果不是顶级分类，则parent_id等于获取的parent_id，parent_id_path = parent_id的path + _id
func AddCategory(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var acSrv AddCateoryService
	if err := c.ShouldBindJSON(&acSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	category := models.Category{}
	category.ComID = claims.ComId
	category.ID = getLastID("category")
	category.CategoryName = acSrv.Name
	category.IsDelete = false
	if acSrv.IsTop {
		category.ParentID = 0
		category.ParentIDPath = "0_" + strconv.Itoa(int(category.ID))
	} else {
		category.ParentID = acSrv.ParentID
		// 先找出父级分类的parentIDPath
		collection := models.Client.Collection("category")
		var res models.Category
		err := collection.FindOne(context.TODO(), bson.D{{"id", acSrv.ParentID}}).Decode(&res)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find parent category",
			})
			return
		}
		category.ParentIDPath = res.ParentIDPath + "_" + strconv.Itoa(int(category.ID))
	}

	collection := models.Client.Collection("category")
	_, err := collection.InsertOne(context.TODO(), category)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't insert category",
		})
		return
	}
	setLastID("category")

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "增加商品分类成功",
	})
}

func ListCategory(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 返回一个分类的列表
	// TODO：有没有更快的算法
	categoryList := make(map[int64]map[string]interface{}) // map[顶级分类]一级分类
	var toplist []int64
	collection := models.Client.Collection("category")

	// 先找出顶级分类，保存为map的键，
	cur, err := collection.Find(context.TODO(), bson.D{{"parent_id", 0}, {"com_id", claims.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find top category",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Category
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode category",
			})
			return
		}
		toplist = append(toplist, res.ID)
		categoryList[res.ID] = make(map[string]interface{})
		categoryList[res.ID]["name"] = res.CategoryName
		categoryList[res.ID]["sub"] = []models.Category{}
	}

	filter := bson.M{}
	if len(toplist) > 0 {
		filter["parent_id"] = bson.M{"$in": toplist}
	}
	filter["com_id"] = claims.ComId
	//option := options.Find()
	//option.Projection = bson.M{"com_id":0, "category_name":0, "parent_id":0, "is_delete":0, "parent_id_path":0, "_id": 0}
	cur, err = collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find sub category",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Category
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode category",
			})
			return
		}
		categoryList[res.ParentID]["sub"] = append(categoryList[res.ParentID]["sub"].([]models.Category), res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "分类列表",
		Data: categoryList,
	})
}

type CategoryService struct {
	CatID   int64  `json:"cat_id"`
	CatName string `json:"cat_name"`
}

// 找出一个分类下的所有商品
func FindOneCategory(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var catSrv CategoryService

	if err := c.ShouldBindJSON(&catSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("category")
	var categories []models.Category
	cur, err := collection.Find(context.TODO(), bson.D{{"parent_id", catSrv.CatID}, {"com_id", claims.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find category",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Category
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode category",
			})
			return
		}
		categories = append(categories, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get products",
		Data: categories,
	})
	return
}

func DeleteCategory(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var catSrv CategoryService

	if err := c.ShouldBindJSON(&catSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	colletion := models.Client.Collection("category")
	_, err := colletion.DeleteOne(context.TODO(), bson.D{{"id", catSrv.CatID}, {"com_id", claims.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't delete category",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete succeed",
	})
	return
}

func UpdateCategory(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var catSrv CategoryService

	if err := c.ShouldBindJSON(&catSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	colletion := models.Client.Collection("category")
	_, err := colletion.UpdateOne(context.TODO(), bson.D{{"id", catSrv.CatID}, {"com_id", claims.ComId}}, bson.M{"$set": bson.M{"category_name": catSrv.CatName}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update category",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update succeed",
	})
	return
}
