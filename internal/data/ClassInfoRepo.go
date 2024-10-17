package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type ClassInfoDBRepo struct {
	data *Data
	log  *log.Helper
}
type ClassInfoCacheRepo struct {
	rdb *redis.Client
	log *log.Helper
}

func NewClassInfoDBRepo(data *Data, logger log.Logger) *ClassInfoDBRepo {
	return &ClassInfoDBRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}
func NewClassInfoCacheRepo(rdb *redis.Client, logger log.Logger) *ClassInfoCacheRepo {
	return &ClassInfoCacheRepo{
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

// SaveManyClassInfosToCache 一次性存多个单个课程信息
func (c ClassInfoCacheRepo) SaveManyClassInfosToCache(ctx context.Context, keys []string, classInfos []*model.ClassInfo) error {
	err := c.rdb.Watch(func(tx *redis.Tx) error {
		// 开始事务
		_, err := tx.TxPipelined(func(pipe redis.Pipeliner) error {
			for k, classInfo := range classInfos {
				val, err := json.Marshal(classInfo)
				if err != nil {
					c.log.Errorw(classLog.Msg, fmt.Sprintf("json Marshal (%v) err", classInfo),
						classLog.Reason, err)
					return err
				}
				// 将数据设置到 Redis 中，使用事务管道
				err = pipe.Set(keys[k], val, Expiration).Err()
				if err != nil {
					c.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:add command(set) to pipe err"),
						classLog.Reason, err)
					return err
				}
			}
			return nil
		})
		return err
	}, keys...) // 监控所有将被设置的键

	if err != nil {
		c.log.Errorw(classLog.Msg, "Redis:watch SaveManyClassInfosToCache",
			classLog.Reason, err)
		return err
	}

	return nil
}

// OnlyAddClassInfosToCache 将整个课表存到缓存中去
func (c ClassInfoCacheRepo) OnlyAddClassInfosToCache(ctx context.Context, key string, classInfos []*model.ClassInfo) error {
	val, err := json.Marshal(classInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("json Marshal (%v) err", classInfos),
			classLog.Reason, err)
		return err
	}
	err = c.rdb.Set(key, val, Expiration).Err()
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Set k(%s)-v(%s)", key, val),
			classLog.Reason, err)
		return err
	}
	return nil
}

// OnlyAddClassInfoToCache 仅添加单个课程信息到缓存中
func (c ClassInfoCacheRepo) OnlyAddClassInfoToCache(ctx context.Context, key string, classInfo *model.ClassInfo) error {
	val, err := json.Marshal(classInfo)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("json Marshal (%v) err", classInfo),
			classLog.Reason, err)
		return err
	}
	err = c.rdb.Set(key, val, Expiration).Err()
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Set k(%s)-v(%s) err", key, val),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (c ClassInfoCacheRepo) GetClassInfosFromCache(ctx context.Context, key string) ([]*model.ClassInfo, error) {
	var classInfos = make([]*model.ClassInfo, 0)
	val, err := c.rdb.Get(key).Result()
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:get key(%s) err", key),
			classLog.Reason, err)
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &classInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("json Unmarshal (%v) err", val),
			classLog.Reason, err)
		return nil, err
	}
	return classInfos, nil
}
func (c ClassInfoCacheRepo) GetClassInfoFromCache(ctx context.Context, key string) (*model.ClassInfo, error) {
	var classInfo = &model.ClassInfo{}
	val, err := c.rdb.Get(key).Result()
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:get key(%s) err", key),
			classLog.Reason, err)
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &classInfo)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("json Unmarshal (%v) err", val),
			classLog.Reason, err)
		return nil, err
	}
	return classInfo, nil
}

// AddClassInfoToCache 添加课程的操作集合
func (c ClassInfoCacheRepo) AddClassInfoToCache(ctx context.Context, classInfoKey, classInfosKey string, classInfo *model.ClassInfo) error {
	oldClassInfos, err := c.GetClassInfosFromCache(ctx, classInfosKey)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:GetClassInfosFromCache(ctx, %s) err", classInfoKey),
			classLog.Reason, err)
		return err
	}
	//将原本的classInfos中要添加的课程添加
	newClassInfos := append(oldClassInfos, classInfo)
	err = c.OnlyAddClassInfosToCache(ctx, classInfosKey, newClassInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:OnlyAddClassInfosToCache(ctx, %s, %v) err", classInfoKey, newClassInfos),
			classLog.Reason, err)
		return err
	}
	//添加单个课程信息到缓存中
	err = c.OnlyAddClassInfoToCache(ctx, classInfoKey, classInfo)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:OnlyAddClassInfoToCache(ctx, %s, %v) err", classInfoKey, classInfo),
			classLog.Reason, err)
		return err
	}
	return nil
}

// FixClassInfoInCache 修改缓存中的课表信息
func (c ClassInfoCacheRepo) FixClassInfoInCache(ctx context.Context, oldID, classInfoKey, classInfosKey string, classInfo *model.ClassInfo) error {
	oldClassInfos, err := c.GetClassInfosFromCache(ctx, classInfosKey)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:GetClassInfosFromCache(ctx, %s) err", classInfoKey),
			classLog.Reason, err)
		return err
	}
	//将原本的classInfos中要更改的课程更改
	for k, oldClassInfo := range oldClassInfos {
		if oldClassInfo.ID == oldID {
			oldClassInfos[k] = classInfo
			break
		}
	}
	err = c.OnlyAddClassInfosToCache(ctx, classInfosKey, oldClassInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:OnlyAddClassInfosToCache(ctx, %s, %v) err", classInfoKey, oldClassInfos),
			classLog.Reason, err)
		return err
	}
	//添加单个课程信息到缓存中
	err = c.OnlyAddClassInfoToCache(ctx, classInfoKey, classInfo)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:OnlyAddClassInfoToCache(ctx, %s, %v) err", classInfoKey, classInfo),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (c ClassInfoCacheRepo) DeleteClassInfoFromCache(ctx context.Context, deletedId, classInfosKey string) error {
	var Indx int
	oldClassInfos, err := c.GetClassInfosFromCache(ctx, classInfosKey)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:GetClassInfosFromCache(ctx, %s) err", classInfosKey),
			classLog.Reason, err)
	}
	for k, oldClassInfo := range oldClassInfos {
		if oldClassInfo.ID == deletedId {
			Indx = k
			break
		}
	}
	newClassInfos := append(oldClassInfos[:Indx], oldClassInfos[Indx+1:]...)
	err = c.OnlyAddClassInfosToCache(ctx, classInfosKey, newClassInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:OnlyAddClassInfosToCache(ctx, %s, %v) err", classInfosKey, newClassInfos),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (c ClassInfoDBRepo) SaveClassInfosToDB(ctx context.Context, classInfo []*model.ClassInfo) error {
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	for _, cla := range classInfo {
		if err := db.FirstOrCreate(cla).Error; err != nil {
			c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create %v in %s", cla, model.ClassInfoTableName),
				classLog.Reason, err)
			return errcode.ErrCourseSave
		}
	}
	return nil
}

func (c ClassInfoDBRepo) AddClassInfoToDB(ctx context.Context, classInfo *model.ClassInfo) error {
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.FirstOrCreate(classInfo).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create %v in %s", classInfo, model.ClassInfoTableName),
			classLog.Reason, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (c ClassInfoDBRepo) GetClassInfoFromDB(ctx context.Context, ID string) (*model.ClassInfo, error) {
	db := c.data.Mysql.Table(model.ClassInfoTableName).WithContext(ctx)
	cla := &model.ClassInfo{}
	err := db.Select([]string{
		"id",
		"jxb_id",
		"day",
		"teacher",
		"where",
		"class_when",
		"week_duration",
		"class_name",
		"credit",
		"weeks",
		"semester",
		"year",
	}).Where("id =?", ID).First(cla).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfo in %s where (id = %s)", model.ClassInfoTableName, ID),
			classLog.Reason, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrClassNotFound
		}
		return nil, errcode.ErrClassFound

	}
	return cla, err
}

func (c ClassInfoDBRepo) DeleteClassInfoInDB(ctx context.Context, ID string) error {
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.Where("id =?", ID).Delete(&model.ClassInfo{}).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:delete class in %s where (id = %s)", model.ClassInfoTableName, ID),
			classLog.Reason, err)
		return errcode.ErrClassDelete
	}
	return nil
}

func (c ClassInfoDBRepo) UpdateClassInfoInDB(ctx context.Context, classInfo *model.ClassInfo) error {
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.Save(classInfo).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:update %v in %s", classInfo, model.ClassInfoTableName),
			classLog.Reason, err)
		return errcode.ErrClassUpdate
	}
	return nil
}
func (c ClassInfoDBRepo) GetAllClassInfos(ctx context.Context, xnm, xqm string) ([]*model.ClassInfo, error) {
	db := c.data.Mysql.Table(model.ClassInfoTableName).WithContext(ctx)
	cla := make([]*model.ClassInfo, 0)
	err := db.Where("year = ? AND semester = ?", xnm, xqm).Find(&cla).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfos in %s where (year = %s,semester = %s)",
			model.ClassInfoTableName, xnm, xqm),
			classLog.Reason, err)
		return nil, err
	}
	return cla, nil
}
