package data

import (
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
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
	Expiration        = 5 * 24 * time.Hour
	RecycleExpiration = 2 * 30 * 24 * time.Hour
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
	NewJxbDBRepo,
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
func NewDB(c *conf.Data, logfile *os.File) *gorm.DB {
	//var logfile *os.File
	//var err error
	//filename := filepath.Join(c.Database.LogPath, c.Database.LogFileName)
	//// 判断日志路径是否存在，如果不存在就创建
	//if exist := tool.IsExist(c.Database.LogPath); !exist {
	//	if err := os.MkdirAll(c.Database.LogPath, os.ModePerm); err != nil {
	//		return nil
	//	}
	//}
	//if exist := tool.IsExist(filename); !exist {
	//	logfile, err = os.Create(filepath.Join(filename))
	//} else {
	//	logfile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//}
	//if err != nil {
	//	panic(err)
	//}
	//defer logfile.Close() // 确保文件在函数退出时关闭
	newlogger := logger2.New(
		logger3.New(logfile, "\r\n", logger3.LstdFlags),
		logger2.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger2.Info,
			Colorful:      false,
		},
	)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{Logger: newlogger})
	if err != nil {
		panic("connect mysql failed")
	}
	if err := db.AutoMigrate(&model.ClassInfo{}, &model.StudentCourse{}, &model.Jxb{}); err != nil {
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
		Password:     c.Redis.Password,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	return rdb
}
