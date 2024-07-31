package data

import (
	"class/internal/biz"
	"class/internal/errcode"
	log2 "class/internal/log"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

type classrepo struct {
	Data *Data
	log  log2.LogerPrinter
}

func NewClassRepo(data *Data, log log2.LogerPrinter) biz.ClassRepo {
	return &classrepo{
		Data: data,
		log:  log,
	}
}
func (c classrepo) SaveClassInfo(ctx context.Context, clas []*biz.ClassInfo) error {
	tx := c.Data.db.Table(biz.ClassInfoTableName).WithContext(ctx).Begin()
	for _, cl := range clas {

		err := tx.Create(&cl).Error()
		if err != nil {
			c.log.FuncError(tx.Create, err)
			tx.Rollback()
			return errcode.ErrCourseSave
		}

		//缓存
		err = c.Data.cache.Set(cl.GetKey(), cl, 24*7*time.Hour)
		if err != nil {
			c.log.FuncError(c.Data.cache.Set, err)
		}
	}

	return tx.Commit()
}

func (c classrepo) AddClassInfo(ctx context.Context, cla *biz.ClassInfo) error {
	db := c.Data.db.Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.Create(cla).Error()
	if err != nil {
		c.log.FuncError(db.Create, err)
		return errcode.ErrClassUpdate
	}
	go func() {
		err = c.Data.cache.Set(cla.GetKey(), cla, 24*7*time.Hour)
		if err != nil {
			c.log.FuncError(c.Data.cache.Set, err)
		}
	}()
	return nil
}

func (c classrepo) GetSpecificClassInfo(ctx context.Context, id string, xnm, xqm string, day int64, dur string) ([]*biz.ClassInfo, error) {
	var classes = make([]*biz.ClassInfo, 0)
	var keys []string
	cacheGet := true

	aimKey := GenerateKeySpecific(id, xnm, xqm, day, dur)
	keys, err := c.Data.cache.ScanKeys(aimKey)
	if err != nil {
		c.log.FuncError(c.Data.cache.ScanKeys, err)
		cacheGet = false
	}

	if len(keys) == 0 {
		cacheGet = false
	} else {
		classes, err = getClassesFromCache(c.Data.cache, keys)
		if err != nil {
			c.log.FuncError(getClassesFromCache, err)
			cacheGet = false
		}
	}

	if !cacheGet {
		db := c.Data.db.Table(biz.ClassInfoTableName).WithContext(ctx)
		classes, err = db.GetSpecificClassInfos(id, xnm, xqm, day, dur)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.log.FuncError(db.GetSpecificClassInfos, err)
				return nil, errcode.ErrClassNotFound
			}
			return nil, errcode.ErrClassFound
		}
	}

	return classes, nil
}
func (c classrepo) GetClasses(ctx context.Context, id string, xnm, xqm string) ([]*biz.ClassInfo, error) {
	db := c.Data.db.Table(biz.ClassInfoTableName).WithContext(ctx)
	var classes = make([]*biz.ClassInfo, 0)
	cacheGet := true
	aimKey := GenerateKeyExtensive(id, xnm, xqm)
	keys, err := c.Data.cache.ScanKeys(aimKey)
	if err != nil {
		c.log.FuncError(c.Data.cache.ScanKeys, err)
		cacheGet = false
	}
	if len(keys) == 0 {
		cacheGet = false
	} else {
		//从缓存获取课程
		classes, err = getClassesFromCache(c.Data.cache, keys)
		if err != nil {
			c.log.FuncError(getClassesFromCache, err)
			cacheGet = false
		}
	}
	// 缓存击穿
	if !cacheGet {
		//从数据库中获取
		classes, err = db.GetClassInfos(id, xnm, xqm)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.log.FuncError(db.GetClassInfos, err)
				return nil, errcode.ErrClassNotFound
			}
			return nil, errcode.ErrClassFound
		}
	}

	return classes, nil
}

func (c classrepo) DeleteClass(ctx context.Context, id string) error {
	db := c.Data.db.Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.DeleteClassInfo(id)
	if err != nil {
		c.log.FuncError(db.DeleteClassInfo, err)
		return errcode.ErrClassDelete
	}
	go func() {
		key, err := ParseID(id)
		if err != nil {
			c.log.FuncError(ParseID, err)
			return
		}
		err = c.Data.cache.DeleteKey(key)
		if err != nil {
			c.log.FuncError(c.Data.cache.DeleteKey, err)
		}
	}()
	return nil
}
func getClassesFromCache(cache Cache, keys []string) ([]*biz.ClassInfo, error) {
	var classes []*biz.ClassInfo

	for _, key := range keys {
		classInfo, err := cache.GetClassInfo(key)
		if err != nil {
			return nil, err
		}
		classes = append(classes, classInfo)
	}
	return classes, nil
}

func GenerateKeySpecific(id string, xnm, xqm string, day int64, dur string) string {
	return fmt.Sprintf("class:%s:%s:%s:%d:%s*", id, xnm, xqm, day, dur)
}
func GenerateKeyExtensive(id string, xnm, xqm string) string {
	return fmt.Sprintf("class:%s:%s:%s*", id, xnm, xqm)
}
func ParseID(id string) (string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 8 {
		return "", fmt.Errorf("invalid ID format")
	}
	return fmt.Sprintf("class:%s:%s:%s:%s:%s:%s", parts[0], parts[2], parts[3], parts[4], parts[5], id), nil
}
