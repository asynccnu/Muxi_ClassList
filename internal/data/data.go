package data

import (
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/cache"
	cmodel "github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/repo"
	"github.com/asynccnu/Muxi_ClassList/internal/data/jxb"
	jmodel "github.com/asynccnu/Muxi_ClassList/internal/data/jxb/model"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	logger2 "gorm.io/gorm/logger"
	logger3 "log"
	"os"
	"time"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedisDB,
	repo.NewClassRepo,
	cache.NewCache,
	jxb.NewJxbDBRepo,
)

// Data .
type Data struct {
	Mysql *gorm.DB
}

// NewData .
func NewData(c *conf.Data, mysqlDB *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		Mysql: mysqlDB,
	}, cleanup, nil
}

// NewDB 连接mysql数据库
func NewDB(c *conf.Data) *gorm.DB {
	//注意:
	//这个logfile 最好别在此处声明,最好在main函数中声明,在程序结束时关闭
	//否则你只能在下面的db.AutoMigrate得到相关日志
	newlogger := logger2.New(
		//日志写入文件
		logger3.New(os.Stdout, "\r\n", logger3.LstdFlags),
		logger2.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger2.Warn,
			Colorful:      false,
		},
	)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{Logger: newlogger})
	if err != nil {
		panic(fmt.Sprintf("connect mysql failed:%v", err))
	}
	if err := db.AutoMigrate(&cmodel.ClassDO{}, &cmodel.StudentClassRelationDO{}, &jmodel.Jxb{}); err != nil {
		panic(fmt.Sprintf("mysql auto migrate failed:%v", err))
	}
	classLog.LogPrinter.Info("connect mysql successfully")
	return db
}

// NewRedisDB 连接redis
func NewRedisDB(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		DB:           0,
		Password:     c.Redis.Password,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("connect redis err:%v", err))
	}
	classLog.LogPrinter.Info("connect redis successfully")
	return rdb
}
