package data

import (
	"class/internal/biz"
	"class/internal/errcode"
	log2 "class/internal/log"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	Expiration = 7 * 24 * time.Hour
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
	tx := c.Data.db.WithContext(ctx).Begin() // 统一事务处理

	// 使用切片收集缓存更新操作
	var cacheOps []func()

	// 处理 ClassInfo
	for _, cl := range clas {
		if err := tx.Table(biz.ClassInfoTableName).Create(&cl).Error(); err != nil {
			c.log.FuncError(tx.Create, err)
			tx.Rollback()
			return errcode.ErrCourseSave
		}
		// 收集缓存操作
		cacheOps = append(cacheOps, func(cl *biz.ClassInfo) func() {
			return func() {
				key := cl.ID
				err := c.Data.cache.Set(key, cl, Expiration)
				if err != nil {
					c.log.FuncError(c.Data.cache.Set, err)
				}
			}
		}(cl))
	}

	// 处理 StudentCourse
	for _, sc := range Scs {
		if err := tx.Table(biz.StudentCourseTableName).Create(&sc).Error(); err != nil {
			c.log.FuncError(tx.Create, err)
			tx.Rollback()
			return errcode.ErrCourseSave
		}
		// 收集缓存操作
		cacheOps = append(cacheOps, func(sc *biz.StudentCourse) func() {
			return func() {
				err := c.Data.cache.AddEleToSet(sc.StuID, sc.Year, sc.Semester, sc.ClaID)
				if err != nil {
					c.log.FuncError(c.Data.cache.AddEleToSet, err)
				}
			}
		}(sc))
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errcode.ErrCourseSave
	}
	go func() {
		// 在事务成功后执行缓存操作
		for _, op := range cacheOps {
			op()
		}
	}()
	return nil
}

func (c classrepo) AddClassInfo(ctx context.Context, cla *biz.ClassInfo, sc *biz.StudentCourse) error {
	tx := c.Data.db.WithContext(ctx).Begin() // 统一事务处理

	// 处理 ClassInfo
	if err := tx.Table(biz.ClassInfoTableName).Create(cla).Error(); err != nil {
		c.log.FuncError(tx.Create, err)
		tx.Rollback()
		return errcode.ErrClassUpdate
	}

	// 处理 StudentCourse
	if err := tx.Table(biz.StudentCourseTableName).Create(sc).Error(); err != nil {
		c.log.FuncError(tx.Create, err)
		tx.Rollback()
		return errcode.ErrClassUpdate
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errcode.ErrClassUpdate
	}

	// 在事务成功提交后，异步处理缓存更新
	go func() {
		// 课程信息缓存
		key := cla.ID
		err := c.Data.cache.Set(key, cla, 7*24)
		if err != nil {
			c.log.FuncError(c.Data.cache.Set, err)
		}
	}()

	go func() {
		// 缓存 StudentCourse
		err := c.Data.cache.AddEleToSet(sc.StuID, sc.Year, sc.Semester, sc.ClaID)
		if err != nil {
			c.log.FuncError(c.Data.cache.AddEleToSet, err)
		}
	}()

	// 不等待缓存写入完成，直接返回
	return nil
}
func (c classrepo) GetSpecificClassInfo(ctx context.Context, stuId string, xnm, xqm string, day int64, dur string) ([]*biz.ClassInfo, error) {
	//TODO:重构
	var classInfos = make([]*biz.ClassInfo, 0)
	cacheGet := true
	//key1 := GenerateSetName(stuId, xnm, xqm)
	claIds, err := c.Data.cache.GetClassIDFromSet(stuId, xnm, xqm)
	if err != nil {
		c.log.FuncError(c.Data.cache.ScanKeys, err)
		cacheGet = false
	}
	if len(claIds) == 0 {
		cacheGet = false
	} else {
		for _, Id := range claIds {
			if !Check(Id, day, dur) {
				continue
			} //筛选符合要求的ID
			key := Id
			classInfo, err := c.Data.cache.GetClassInfo(key)
			if err != nil {
				c.log.FuncError(c.Data.cache.GetClassInfo, err)
				cacheGet = false
			}
			classInfos = append(classInfos, classInfo)
		}
	}
	//缓存获取失败
	if !cacheGet {
		classIds, err := c.Data.db.GetClassIds(ctx, stuId)
		if err != nil {
			c.log.FuncError(c.Data.db.GetClassIds, err)
			return nil, errcode.ErrClassNotFound
		}

		for _, Id := range classIds {
			if !Check(Id, day, dur) {
				continue
			} //筛选符合要求的ID
			classInfo, err := c.Data.db.GetSpecificClassInfos(ctx, Id)
			if err != nil {
				c.log.FuncError(c.Data.db.GetSpecificClassInfos, err)
				return nil, errcode.ErrClassNotFound
			}
			classInfos = append(classInfos, classInfo)
		}
	}
	return classInfos, nil

}
func (c classrepo) GetClasses(ctx context.Context, stuId string, xnm, xqm string) ([]*biz.ClassInfo, error) {
	var classInfos = make([]*biz.ClassInfo, 0)
	cacheGet := true
	//key1 := GenerateSetName(stuId, xnm, xqm)
	claIds, err := c.Data.cache.GetClassIDFromSet(stuId, xnm, xqm)
	if err != nil {
		c.log.FuncError(c.Data.cache.ScanKeys, err)
		cacheGet = false
	}
	if len(claIds) == 0 {
		cacheGet = false
	} else {
		for _, classId := range claIds {
			key := classId
			classInfo, err := c.Data.cache.GetClassInfo(key)
			if err != nil {
				c.log.FuncError(c.Data.cache.GetClassInfo, err)
				cacheGet = false
			}
			classInfos = append(classInfos, classInfo)
		}
	}
	//缓存获取失败
	if !cacheGet {
		classIds, err := c.Data.db.GetClassIds(ctx, stuId)
		if err != nil {
			c.log.FuncError(c.Data.db.GetClassIds, err)
			return nil, errcode.ErrClassNotFound
		}

		for _, Id := range classIds {
			classInfo, err := c.Data.db.GetSpecificClassInfos(ctx, Id)
			if err != nil {
				c.log.FuncError(c.Data.db.GetSpecificClassInfos, err)
				return nil, errcode.ErrClassNotFound
			}
			classInfos = append(classInfos, classInfo)
		}
	}
	return classInfos, nil
}

func (c classrepo) DeleteClass(ctx context.Context, id string) error {
	//TODO:
}

func Check(id string, day int64, dur string) bool {
	day1, dur1, err := ExtractDayAndClassWhen(id)
	if err != nil {
		return false
	}
	if day != day1 || dur != dur1 {
		return false
	}
	return true
}

// ExtractDayAndClassWhen 提取格式化字符串中的 day 和 classwhen
func ExtractDayAndClassWhen(id string) (int64, string, error) {
	// 定义正则表达式来匹配 day 和 classwhen
	re := regexp.MustCompile(`^Class:\w+:\w+:\w+:(\d+):(\w+):`)

	// 找到匹配的子字符串
	matches := re.FindStringSubmatch(id)
	if len(matches) < 3 {
		return 0, "", fmt.Errorf("could not extract day and classwhen from ID: %s", id)
	}

	// 将 day 转换为 int
	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("error converting day to int: %v", err)
	}

	classwhen := matches[2]
	return int64(day), classwhen, nil
}
