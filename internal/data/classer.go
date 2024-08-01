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
	"sync"
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
func (c classrepo) SaveClassInfo(ctx context.Context, clas []*biz.ClassInfo, Scs []*biz.StudentCourse) error {
	var err1, err2 error
	tx := c.Data.db.WithContext(ctx).Begin() // 统一事务处理

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// 处理 ClassInfo
	go func() {
		defer wg.Done()
		for _, cl := range clas {
			if err := tx.Table(biz.ClassInfoTableName).Create(&cl).Error(); err != nil {
				c.log.FuncError(tx.Create, err)
				err1 = err // 记录错误
				return
			}
			// TODO: 缓存
		}
	}()

	// 处理 StudentCourse
	go func() {
		defer wg.Done()
		for _, sc := range Scs {
			if err := tx.Table(biz.StudentCourseTableName).Create(&sc).Error(); err != nil {
				c.log.FuncError(tx.Create, err)
				err2 = err // 记录错误
				return
			}
			// TODO: 缓存
		}
	}()

	wg.Wait()

	// 在 Wait 之后检查错误，如果存在任何一个错误则回滚
	if err1 != nil || err2 != nil {
		tx.Rollback()
		return errcode.ErrCourseSave
	}

	// 如果没有错误，提交事务
	if err := tx.Commit(); err != nil {
		return errcode.ErrCourseSave
	}

	return nil
}

func (c classrepo) AddClassInfo(ctx context.Context, cla *biz.ClassInfo, sc *biz.StudentCourse) error {
	var err1, err2 error
	tx := c.Data.db.WithContext(ctx).Begin() // 统一事务处理

	// 使用 sync.WaitGroup 等待两个并发操作完成
	var wg sync.WaitGroup
	wg.Add(2)

	// 处理 ClassInfo
	go func() {
		defer wg.Done()
		if err := tx.Table(biz.ClassInfoTableName).Create(cla).Error(); err != nil {
			c.log.FuncError(tx.Create, err)
			err1 = err // 记录错误
			return
		}
		// TODO: 缓存
	}()

	// 处理 StudentCourse
	go func() {
		defer wg.Done()
		if err := tx.Table(biz.StudentCourseTableName).Create(sc).Error(); err != nil {
			c.log.FuncError(tx.Create, err)
			err2 = err // 记录错误
			return
		}
		// TODO: 缓存
	}()

	// 等待所有 goroutine 完成
	wg.Wait()

	// 检查错误并进行回滚
	if err1 != nil || err2 != nil {
		tx.Rollback()
		return errcode.ErrClassUpdate
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errcode.ErrClassUpdate
	}

	return nil
}

func (c classrepo) GetSpecificClassInfo(ctx context.Context, id string, xnm, xqm string, day int64, dur string) ([]*biz.ClassInfo, error) {
	//TODO:重构
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
	//TODO:重构

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
