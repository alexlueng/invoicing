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
	Comment  string `json:"comment"`   // 备注
	//Thumbnail models.Image `json:"thumbnail"` // 图片
	URLs []CategoryImageURL `json:"urls"`
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
	category.ID = GetLastID("category")
	category.CategoryName = acSrv.Name
	category.Comment = acSrv.Comment
	category.IsDelete = false
	if acSrv.IsTop {
		category.ParentID = 0
		category.ParentIDPath = "0_" + strconv.Itoa(int(category.ID))
	} else {
		category.ParentID = acSrv.ParentID
		// 先找出父级分类的parentIDPath
		res, err := models.SelectCategoryById(acSrv.ParentID)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find parent category",
			})
			return
		}
		category.ParentIDPath = res.ParentIDPath + "_" + strconv.Itoa(int(category.ID))
		category.Level = res.Level + 1
	}

	err := category.Add()
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't insert category",
		})
		return
	}
	SetLastID("category")

	// 将图片保存到分类图片表中
	collection := models.Client.Collection("category_image")
	for _, url := range acSrv.URLs {
		catImage := models.CategoryImage{
			ComID:      claims.ComId,
			CategoryID: category.ID,
			ImageID:    GetLastID("category_image"),
			LocalPath:  url.LocalURL,
			CloudPath:  url.CloudURL,
			IsDelete:   false,
		}
		_, err := collection.InsertOne(context.TODO(), catImage)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Insert category image error",
			})
			return
		}
		SetLastID("category_image")
	}

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

	// 先找出顶级分类，保存为map的键
	topCategoryList, err := models.SelectTopCategoryByComId(claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find top category",
		})
		return
	}

	for _, res := range topCategoryList.Category {
		toplist = append(toplist, res.ID)
		categoryList[res.ID] = make(map[string]interface{})
		categoryList[res.ID]["name"] = res.CategoryName
		categoryList[res.ID]["sub"] = []models.Category{}
	}

	subCategoryList, err := models.SelectCategoryByComIdAndTopCategoryID(claims.ComId, toplist)
	if err != nil {
		if err.Error() != "未获取到分类数据" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find sub category",
			})
			return
		}
	}

	if subCategoryList != nil {
		for _, res := range subCategoryList.Category {
			if categoryList[res.ParentID]["sub"] != nil {
				categoryList[res.ParentID]["sub"] = append(categoryList[res.ParentID]["sub"].([]models.Category), res)
			}
		}
	}

	// 返回分类图片
	collection := models.Client.Collection("category_image")
	var images []models.CategoryImage

	cur, err := collection.Find(context.TODO(), bson.D{{"com_id", claims.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find category images",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.CategoryImage
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode category image",
			})
			return
		}
		images = append(images, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "分类列表",
		Data: map[string]interface{}{
			"category": categoryList,
			"images":   images,
		},
	})
}

type CategoryService struct {
	CatID   int64              `json:"cat_id"`
	CatName string             `json:"cat_name"`
	Comment string             `json:"comment"` // 备注
	URLs    []CategoryImageURL `json:"urls"`
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

	category, err := models.SelectCategoryByComIdAndParentId(catSrv.CatID, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find category",
		})
		return
	}

	var categories []models.Category
	for _, res := range category.Category {
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

	err := models.DeleteCategoryByComIdAndId(claims.ComId, catSrv.CatID)
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

	err := models.UpdateCategory(claims.ComId, catSrv.CatID, catSrv.CatName)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update category",
		})
		return
	}

	collection := models.Client.Collection("category_image")
	for _, url := range catSrv.URLs {
		if url.CategoryID == 0 {
			image := models.CategoryImage{
				ComID:     claims.ComId,
				CategoryID: catSrv.CatID,
				ImageID:   GetLastID("category_image"),
				LocalPath: url.LocalURL,
				CloudPath: url.CloudURL,
				IsDelete:  false,
			}
			_, err := collection.InsertOne(context.TODO(), image)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't insert image",
				})
				return
			}
			SetLastID("category_image")
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update succeed",
	})
	return
}
