package data

import (
	"class/internal/biz"
	"class/internal/errcode"
	log "class/internal/logPrinter"
	"context"
	"encoding/json"
	"errors"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type ClassInfoDBRepo struct {
	data *Data
	log  log.LogerPrinter
}
type ClassInfoCacheRepo struct {
	rdb *redis.Client
	log log.LogerPrinter
}

func NewClassInfoDBRepo(data *Data, log log.LogerPrinter) *ClassInfoDBRepo {
	return &ClassInfoDBRepo{
		log:  log,
		data: data,
	}
}
func NewClassInfoCacheRepo(rdb *redis.Client, log log.LogerPrinter) *ClassInfoCacheRepo {
	return &ClassInfoCacheRepo{
		rdb: rdb,
		log: log,
	}
}

func (c ClassInfoCacheRepo) SaveManyClassInfosToCache(ctx context.Context, keys []string, classInfos []*biz.ClassInfo) error {
	err := c.rdb.Watch(func(tx *redis.Tx) error {
		// 开始事务
		_, err := tx.TxPipelined(func(pipe redis.Pipeliner) error {
			for k, classInfo := range classInfos {
				val, err := json.Marshal(classInfo)
				if err != nil {
					return err
				}
				// 将数据设置到 Redis 中，使用事务管道
				err = pipe.Set(keys[k], val, Expiration).Err()
				if err != nil {
					return err
				}
			}
			return nil
		})
		return err
	}, keys...) // 监控所有将被设置的键

	if err != nil {
		c.log.FuncError(c.rdb.Watch, err)
		return err
	}
	return nil
}
func (c ClassInfoCacheRepo) AddClassInfoToCache(ctx context.Context, key string, classInfo *biz.ClassInfo) error {
	val, err := json.Marshal(classInfo)
	if err != nil {
		c.log.FuncError(json.Marshal, err)
		return err
	}
	err = c.rdb.Set(key, val, Expiration).Err()
	if err != nil {
		c.log.FuncError(c.rdb.Set, err)
		return err
	}
	return nil
}

func (c ClassInfoCacheRepo) GetClassInfoFromCache(ctx context.Context, key string) (*biz.ClassInfo, error) {
	var classInfo = &biz.ClassInfo{}
	val, err := c.rdb.Get(key).Result()
	if err != nil {
		c.log.FuncError(c.rdb.Get, err)
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &classInfo)
	if err != nil {
		c.log.FuncError(json.Unmarshal, err)
		return nil, err
	}
	return classInfo, nil
}

func (c ClassInfoCacheRepo) DeleteClassInfoFromCache(ctx context.Context, key string) error {
	err := c.rdb.Del(key).Err()
	if err != nil {
		c.log.FuncError(c.rdb.Del, err)
		return err
	}
	return nil
}

func (c ClassInfoCacheRepo) UpdateClassInfoInCache(ctx context.Context, key string, classInfo *biz.ClassInfo) error {
	val, err := json.Marshal(classInfo)
	if err != nil {
		c.log.FuncError(json.Marshal, err)
		return err
	}
	err = c.rdb.Set(key, val, Expiration).Err()
	if err != nil {
		c.log.FuncError(c.rdb.Set, err)
		return err
	}
	return nil
}

func (c ClassInfoDBRepo) SaveClassInfosToDB(ctx context.Context, classInfo []*biz.ClassInfo) error {
	db := c.data.DB(ctx).Table(biz.ClassInfoTableName).WithContext(ctx)
	for _, cla := range classInfo {
		if err := db.FirstOrCreate(cla).Error; err != nil {
			c.log.FuncError(db.Create, err)
			return errcode.ErrCourseSave

		}
	}
	return nil
}

func (c ClassInfoDBRepo) AddClassInfoToDB(ctx context.Context, classInfo *biz.ClassInfo) error {
	db := c.data.DB(ctx).Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.FirstOrCreate(classInfo).Error
	if err != nil {
		c.log.FuncError(db.Create, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (c ClassInfoDBRepo) GetClassInfoFromDB(ctx context.Context, ID string) (*biz.ClassInfo, error) {
	db := c.data.Mysql.Table(biz.ClassInfoTableName).WithContext(ctx)
	cla := &biz.ClassInfo{}
	err := db.Where("id =?", ID).First(cla).Error
	if err != nil {

		c.log.FuncError(db.Where, err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrClassNotFound
		}
		return nil, errcode.ErrClassFound

	}
	return cla, err
}

func (c ClassInfoDBRepo) DeleteClassInfoInDB(ctx context.Context, ID string) error {
	db := c.data.DB(ctx).Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.Where("id =?", ID).Delete(&biz.ClassInfo{}).Error
	if err != nil {
		c.log.FuncError(db.Where, err)
		return errcode.ErrClassDelete
	}
	return nil
}

func (c ClassInfoDBRepo) UpdateClassInfoInDB(ctx context.Context, classInfo *biz.ClassInfo) error {
	db := c.data.DB(ctx).Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.Save(classInfo).Error
	if err != nil {
		c.log.FuncError(db.Save, err)
		return errcode.ErrClassUpdate
	}
	return nil
}
