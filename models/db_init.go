package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	//Client     *mongo.Client
	Collection *mongo.Collection
	// 数据库
	DataBase      *mongo.Database
	insertOneRes  *mongo.InsertOneResult
	insertManyRes *mongo.InsertManyResult
	delRes        *mongo.DeleteResult
	updateRes     *mongo.UpdateResult
	cursor        *mongo.Cursor
	//howieArray      = GetHowieArray()
	//howie           Howie
	//howieArrayEmpty []Howie
	size int64
)

// 数据库连接
func ConnDb(dbUrl string) {
	//want, err := readpref.New(readpref.SecondaryMode) //表示只使用辅助节点
	//if err != nil {
	//	checkErr(err)
	//}
	//wc := writeconcern.New(writeconcern.WMajority())
	//readconcern.Majority()
	////链接mongo服务
	//opt := options.Client().ApplyURI(dbUrl)
	//opt.SetLocalThreshold(3 * time.Second)     //只使用与mongo操作耗时小于3秒的
	//opt.SetMaxConnIdleTime(5 * time.Second)    //指定连接可以保持空闲的最大毫秒数
	//opt.SetMaxPoolSize(200)                    //使用最大的连接数
	//opt.SetReadPreference(want)                //表示只使用辅助节点
	//opt.SetReadConcern(readconcern.Majority()) //指定查询应返回实例的最新数据确认为，已写入副本集中的大多数成员
	//opt.SetWriteConcern(wc)                    //请求确认写操作传播到大多数mongod实例
	//if Client, err = mongo.Connect(getContext(), opt); err != nil {
	//	checkErr(err)
	//}
	////UseSession(client)
	////判断服务是否可用
	//if err = Client.Ping(getContext(), readpref.Primary()); err != nil {
	//	checkErr(err)
	//}
	//
	//DataBase = Client.Database("invoicing")
	//
	////选择数据库和集合
	//Collection = Client.Database("testing_base").Collection("howie")
	//
	////删除这个集合
	//Collection.Drop(getContext())
	//
	////插入一条数据
	//if insertOneRes, err = Collection.InsertOne(getContext(), howieArray[0]); err != nil {
	//	checkErr(err)
	//}
	//
	//fmt.Printf("InsertOne插入的消息ID:%v\n", insertOneRes.InsertedID)
	//
	////批量插入数据
	//if insertManyRes, err = Collection.InsertMany(getContext(), howieArray[1:]); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("InsertMany插入的消息ID:%v\n", insertManyRes.InsertedIDs)
	//var Dinfo = make(map[string]interface{})
	//err = Collection.FindOne(getContext(), bson.D{{"name", "howie_2"}, {"age", 11}}).Decode(&Dinfo)
	//fmt.Println(Dinfo)
	//fmt.Println("_id", Dinfo["_id"])
	//
	////查询单条数据
	//if err = Collection.FindOne(getContext(), bson.D{{"name", "howie_2"}, {"age", 11}}).Decode(&howie); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("FindOne查询到的数据:%v\n", howie)
	//
	////查询单条数据后删除该数据
	//if err = Collection.FindOneAndDelete(getContext(), bson.D{{"name", "howie_3"}}).Decode(&howie); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("FindOneAndDelete查询到的数据:%v\n", howie)
	//
	////查询单条数据后修改该数据
	//if err = Collection.FindOneAndUpdate(getContext(), bson.D{{"name", "howie_4"}}, bson.M{"$set": bson.M{"name": "这条数据我需要修改了"}}).Decode(&howie); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("FindOneAndUpdate查询到的数据:%v\n", howie)
	//
	////查询单条数据后替换该数据(以前的数据全部清空)
	//if err = Collection.FindOneAndReplace(getContext(), bson.D{{"name", "howie_5"}}, bson.M{"hero": "这条数据我替换了"}).Decode(&howie); err != nil {
	//	checkErr(err)
	//}
	//
	//fmt.Printf("FindOneAndReplace查询到的数据:%v\n", howie)
	////一次查询多条数据
	//// 查询createtime>=3
	//// 限制取2条
	//// createtime从大到小排序的数据
	//if cursor, err = Collection.Find(getContext(), bson.M{"createtime": bson.M{"$gte": 2}}, options.Find().SetLimit(2), options.Find().SetSort(bson.M{"createtime": -1})); err != nil {
	//	checkErr(err)
	//}
	//if err = cursor.Err(); err != nil {
	//	checkErr(err)
	//}
	//defer cursor.Close(context.Background())
	//for cursor.Next(context.Background()) {
	//	if err = cursor.Decode(&howie); err != nil {
	//		checkErr(err)
	//	}
	//	howieArrayEmpty = append(howieArrayEmpty, howie)
	//}
	//for _, v := range howieArrayEmpty {
	//	fmt.Printf("Find查询到的数据ObejectId值%s 值:%v\n", v.HowieId.Hex(), v)
	//}
	////查询集合里面有多少数据
	//if size, err = Collection.CountDocuments(getContext(), bson.D{}); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("Count里面有多少条数据:%d\n", size)
	//
	////查询集合里面有多少数据(查询createtime>=3的数据)
	//if size, err = Collection.CountDocuments(getContext(), bson.M{"createtime": bson.M{"$gte": 3}}); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("Count里面有多少条数据:%d\n", size)
	//
	////修改一条数据
	//if updateRes, err = Collection.UpdateOne(getContext(), bson.M{"name": "howie_2"}, bson.M{"$set": bson.M{"name": "我要改了他的名字"}}); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("UpdateOne的数据:%d\n", updateRes)
	//
	////修改多条数据
	//if updateRes, err = Collection.UpdateMany(getContext(), bson.M{"createtime": bson.M{"$gte": 3}}, bson.M{"$set": bson.M{"name": "我要批量改了他的名字"}}); err != nil {
	//	checkErr(err)
	//}
	//
	//fmt.Printf("UpdateMany的数据:%d\n", updateRes)
	////删除一条数据
	//if delRes, err = Collection.DeleteOne(getContext(), bson.M{"name": "howie_1"}); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("DeleteOne删除了多少条数据:%d\n", delRes.DeletedCount)
	//
	////删除多条数据
	//if delRes, err = Collection.DeleteMany(getContext(), bson.M{"createtime": bson.M{"$gte": 7}}); err != nil {
	//	checkErr(err)
	//}
	//fmt.Printf("DeleteMany删除了多少条数据:%d\n", delRes.DeletedCount)
}

func getContext() (ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	return
}

func checkErr(e error) {

}

//package models
//
//import (
//	"invoicing/util"
//	"time"
//
//	"github.com/jinzhu/gorm"
//
//	//
//	_ "github.com/jinzhu/gorm/dialects/mysql"
//)
//
//// DB 数据库链接单例
//var DB *gorm.DB
//
//// Database 在中间件中初始化mysql链接
//func Database(connString string) {
//	db, err := gorm.Open("mysql", connString)
//	db.LogMode(true)
//	// Error
//	if err != nil {
//		util.Log().Panic("连接数据库不成功", err)
//	}
//	//设置连接池
//	//空闲
//	db.DB().SetMaxIdleConns(50)
//	//打开
//	db.DB().SetMaxOpenConns(100)
//	//超时
//	db.DB().SetConnMaxLifetime(time.Second * 30)
//
//	DB = db
//
//	migration()
//}
