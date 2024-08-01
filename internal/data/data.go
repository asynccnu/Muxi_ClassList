package data

import (
	"class/internal/biz"
	"class/internal/conf"
	"context"
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
var ProviderSet = wire.NewSet(NewData, NewClassRepo, NewDB, NewRedisDB, NewGormDatabase, NewRedisCache)

// Data .
type Data struct {
	db    Database
	cache Cache
}
type Database interface {
	Begin() Database
	Create(value interface{}) Database
	WithContext(ctx context.Context) Database
	Table(name string) Database
	Commit() error
	Rollback()
	Error() error
	GetClassInfos(id, xnm, xqm string) ([]*biz.ClassInfo, error)
	GetSpecificClassInfos(id string, xnm, xqm string, day int64, dur string) ([]*biz.ClassInfo, error)
	DeleteClassInfo(id string) error
}
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Scan(cursor uint64, match string, count int64) ([]string, uint64, error)
	ScanKeys(pattern string) ([]string, error)
	GetClassInfo(key string) (*biz.ClassInfo, error)
	DeleteKey(key string) error
	AddEleToZset(stu_id string, cla_id string, day, st, end int64) error
	GetClassIDFromZset(stuId string) ([]string, error)
}

// NewData .
func NewData(db Database, cache Cache, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:    db,
		cache: cache,
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
