package models

import (
	"context"
	"jxc/util"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Database

func Database(connString, dbname string) {
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		util.Log().Panic("连接mongodb不成功", err)
	}
	db := client.Database(dbname)
	Client = db
	util.Log().Info(db.Name())
}

// // DB 数据库链接单例
// var DB *gorm.DB

// // Database 在中间件中初始化mysql链接
// func Database(connString string) {
// 	db, err := gorm.Open("mysql", connString)
// 	db.LogMode(true)
// 	// Error
// 	if err != nil {
// 		util.Log().Panic("连接数据库不成功", err)
// 	}
// 	//设置连接池
// 	//空闲
// 	db.DB().SetMaxIdleConns(50)
// 	//打开
// 	db.DB().SetMaxOpenConns(100)
// 	//超时
// 	db.DB().SetConnMaxLifetime(time.Second * 30)

// 	DB = db

// 	migration()
// }
