package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClassInfoDBRepo struct {
	data *Data
	log  classLog.Clogger
}
type ClassInfoCacheRepo struct {
	rdb *redis.Client
	log classLog.Clogger
}

func NewClassInfoDBRepo(data *Data, logger classLog.Clogger) *ClassInfoDBRepo {
	return &ClassInfoDBRepo{
		log:  logger,
		data: data,
	}
}
func NewClassInfoCacheRepo(rdb *redis.Client, logger classLog.Clogger) *ClassInfoCacheRepo {
	return &ClassInfoCacheRepo{
		rdb: rdb,
		log: logger,
	}
}

// AddClaInfosToCache 将整个课表转换成json格式，然后存到缓存中去
func (c ClassInfoCacheRepo) AddClaInfosToCache(ctx context.Context, key string, classInfos []*model.ClassInfo) error {
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

func (c ClassInfoCacheRepo) GetClassInfosFromCache(ctx context.Context, key string) ([]*model.ClassInfo, error) {
	var classInfos = make([]*model.ClassInfo, 0)
	val, err := c.rdb.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("error getting class info from cache: %w", err)
		}
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

// UpdateClassInfoInCache 修改缓存中的课表信息
func (c ClassInfoCacheRepo) UpdateClassInfoInCache(ctx context.Context, oldID, classInfosKey string, classInfo *model.ClassInfo, add bool) error {
	oldClassInfos, err := c.GetClassInfosFromCache(ctx, classInfosKey)
	if oldClassInfos == nil {
		c.log.Warn(classLog.Msg, fmt.Sprintf("func:GetClassInfosFromCache(ctx, %s) get classinfos is empty", classInfosKey))
		return nil
	}
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:GetClassInfosFromCache(ctx, %s) err", classInfosKey),
			classLog.Reason, err)
		return err
	}
	if !add {
		//将原本的classInfos中要更改的课程更改
		for k, oldClassInfo := range oldClassInfos {
			if oldClassInfo.ID == oldID {
				oldClassInfos[k] = classInfo
				break
			}
		}
	} else {
		oldClassInfos = append(oldClassInfos, classInfo)
	}

	err = c.AddClaInfosToCache(ctx, classInfosKey, oldClassInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:AddClaInfosToCache(ctx, %s, %v) err", classInfosKey, oldClassInfos),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (c ClassInfoCacheRepo) DeleteClassInfoFromCache(ctx context.Context, deletedId, classInfosKey string) error {
	var Indx int
	oldClassInfos, err := c.GetClassInfosFromCache(ctx, classInfosKey)
	if errors.Is(err, redis.Nil) || oldClassInfos == nil {
		return nil
	}
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
	err = c.AddClaInfosToCache(ctx, classInfosKey, newClassInfos)
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("func:AddClaInfosToCache(ctx, %s, %v) err", classInfosKey, newClassInfos),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (c ClassInfoDBRepo) SaveClassInfosToDB(ctx context.Context, classInfos []*model.ClassInfo) error {
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&classInfos).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create %v in %s", classInfos, model.ClassInfoTableName),
			classLog.Reason, err)
		return err
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
func (c ClassInfoDBRepo) GetClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*model.ClassInfo, error) {
	db := c.data.Mysql.WithContext(ctx)
	var (
		cla = make([]*model.ClassInfo, 0)
		err error
	)
	if stuId != "" {
		err = db.Raw("SELECT c.id,c.jxb_id,c.day,c.teacher,c.where,c.class_when,c.week_duration,c.class_name,c.credit,c.weeks,c.year,c.semester FROM class_info c WHERE c.id IN (SELECT s.cla_id FROM student_course s WHERE s.stu_id = ? AND s.year = ? AND s.semester = ?) ORDER BY c.day ASC,c.class_when ASC", stuId, xnm, xqm).Scan(&cla).Error
		if err != nil {
			c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfos  where (stu_id = %s,year = %s,semester = %s)",
				stuId, xnm, xqm),
				classLog.Reason, err)
			return nil, err
		}
	} else {
		err = db.Raw("SELECT c.id,c.day,c.teacher,c.where,c.class_when,c.week_duration,c.class_name,c.credit,c.weeks,c.year,c.semester FROM class_info c WHERE c.id IN (SELECT s.cla_id FROM student_course s WHERE s.is_manually_added = ? AND s.year = ? AND s.semester = ? ) ORDER BY c.day ASC,c.class_when ASC", false, xnm, xqm).Scan(&cla).Error
		if err != nil {
			c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfos  where (is_manually_added = %v,year = %s,semester = %s)",
				false, xnm, xqm),
				classLog.Reason, err)
			return nil, err
		}
	}

	return cla, nil
}
