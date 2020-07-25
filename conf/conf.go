package conf

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"jxc/models"
	"jxc/util"
)

var (
	IdWorker *util.Worker
)


// Init 初始化配置项
func Init() {
	// 从本地读取环境变量
	// 使用默认配置路径 .env 读取不到配置文件
	// 所以在这里先获取到当前项目目录，拼接出 .env 的绝对路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load(dir + "/.env")
	godotenv.Load() // .env

	// 设置日志级别
	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	fmt.Println("LOG_LEVEL :", os.Getenv("LOG_LEVEL"))

	// 读取翻译文件
	// if err := LoadLocales("conf/locales/zh-cn.yaml"); err != nil {
	// 	util.Log().Panic("翻译文件加载失败", err)
	// }

	// 连接数据库
	// models.Database(os.Getenv("MYSQL_DSN"))
	models.Database(os.Getenv("URI"), os.Getenv("DBNAME"))
	//	cache.Redis()

	worker, err := util.NewWorker(0)
	if err != nil {
		util.Log().Panic("error while new snowflake worker")
	}

	IdWorker = worker
}
