package data

import (
	"class/internal/biz"
	"class/internal/conf"
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

const (
	Expiration = 7 * 24 * time.Hour
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedisDB,
	NewStudentAndCourseDBRepo,
	NewStudentAndCourseCacheRepo,
	NewClassInfoDBRepo,
	NewClassInfoCacheRepo,
	NewTransaction,
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
	newlogger := logger2.New(
		logger3.New(os.Stdout, "\r\n", logger3.LstdFlags),
		logger2.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger2.Info,
			Colorful:      true,
		},
	)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{Logger: newlogger})
	if err != nil {
		panic("connect mysql failed")
	}
	if err := db.AutoMigrate(&biz.ClassInfo{}, &biz.StudentCourse{}); err != nil {
		panic(err)
	}
	return db
}

// NewRedisDB 连接redis
func NewRedisDB(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		DB:           0,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	return rdb
}
