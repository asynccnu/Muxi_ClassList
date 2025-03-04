package repo

import (
	"context"
	"errors"
	"fmt"
	bizmodel "github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ClassRepo struct {
	cache      ClassCache
	recycleBin RecycleBinCache
	db         *gorm.DB
}

func NewClassRepo(cache ClassCache, recycleBin RecycleBinCache, db *gorm.DB) *ClassRepo {
	return &ClassRepo{
		cache:      cache,
		recycleBin: recycleBin,
		db:         db,
	}
}

func (c *ClassRepo) checkSCIdsExist(ctx context.Context, stuID, year, semester string, classID string) bool {
	var cnt int64
	err := c.db.Table(model.StudentClassRelationDOTableName).Where("stu_id = ? AND year = ? AND semester = ? AND cla_id = ?", stuID, year, semester, classID).
		Count(&cnt).Error
	if err != nil {
		return false
	}
	return cnt > 0
}

func (c *ClassRepo) GetAllSchoolClassInfos(ctx context.Context, year, semester string, cursor time.Time) []*bizmodel.ClassBiz {
	var classes []*model.ClassDO

	err := c.db.Table(model.ClassDOTableName).
		Where(fmt.Sprintf(
			`%s.year = ? AND %s.semester = ? AND %s.created_at > ?`, model.ClassDOTableName, model.ClassDOTableName, model.ClassDOTableName),
			year, semester, cursor,
		).
		Order(fmt.Sprintf(
			"%s.created_at ASC", model.ClassDOTableName,
		)).
		Limit(100). //最多100个
		Find(&classes).Error
	if err != nil {
		return nil
	}
	return batchNewBizClasses(classes)
}

func (c *ClassRepo) AddClass(ctx context.Context, stuID, year, semester string, class *bizmodel.ClassBiz) error {
	classDO := model.NewClass(class)
	class.ID = classDO.ID

	if c.checkSCIdsExist(ctx, stuID, year, semester, classDO.ID) {
		return errcode.ErrClassIsExist
	}

	err := c.cache.DeleteClassIDList(ctx, stuID, year, semester)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
		return err
	}

	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(classDO).Error
		if err != nil {
			return err
		}
		err = c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.StudentClassRelationDO{
			StuID:           stuID,
			ClaID:           classDO.ID,
			Semester:        semester,
			Year:            year,
			IsManuallyAdded: true,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		classLog.LogPrinter.Errorf("failed to add class[%v] in db: %v", class, err)
		return err
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			err := c.cache.DeleteClassIDList(context.Background(), stuID, year, semester)
			if err != nil {
				classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
			}
		})
	}()
	return nil
}

func (c *ClassRepo) GetAddedClasses(ctx context.Context, stuID, year, semester string) ([]*bizmodel.ClassBiz, error) {
	var classes []*model.ClassDO

	err := c.db.WithContext(ctx).Table(model.ClassDOTableName).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.id = %s.cla_id", model.StudentClassRelationDOTableName, model.ClassDOTableName, model.StudentClassRelationDOTableName)).
		Where(fmt.Sprintf(
			`%s.stu_id = ? AND %s.year = ? AND %s.semester = ? AND %s.is_manually_added = ?`, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName),
			stuID, year, semester, true,
		).Find(&classes).Error
	if err != nil {
		return nil, err
	}

	return batchNewBizClasses(classes), err
}

func (c *ClassRepo) GetRecycledClasses(ctx context.Context, stuID, year, semester string) ([]*bizmodel.ClassBiz, error) {
	ids, err := c.recycleBin.GetRecycledClassIDs(ctx, stuID, year, semester)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	classes, err := c.getClassesFromDBByIDs(ctx, ids...)
	if err != nil {
		return nil, err
	}
	return batchNewBizClasses(classes), nil
}

func (c *ClassRepo) RecoverClassFromRecycledBin(ctx context.Context, stuID, year, semester string, classID string) error {
	isManuallyAdded, err := c.recycleBin.RemoveClassIDFromRecycleBin(ctx, stuID, year, semester, classID)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to remove class_id[%v %v %v %v] from recycle_bin: %v", stuID, year, semester, classID, err)
		return err
	}

	err = c.db.WithContext(ctx).Table(model.StudentClassRelationDOTableName).Create(&model.StudentClassRelationDO{
		StuID:           stuID,
		ClaID:           classID,
		Year:            year,
		Semester:        semester,
		IsManuallyAdded: isManuallyAdded,
	}).Error
	if err != nil {
		classLog.LogPrinter.Errorf("failed to add student_class_relation: %v", err)
		return err
	}

	return nil
}

func (c *ClassRepo) CheckClassIdIsInRecycledBin(ctx context.Context, stuID, year, semester string, classID string) bool {
	return c.recycleBin.CheckRecycleBinElementExist(ctx, stuID, year, semester, classID)
}

func (c *ClassRepo) SaveClass(ctx context.Context, stuID, year, semester string, classes []*bizmodel.ClassBiz) error {
	err := c.cache.DeleteClassIDList(ctx, stuID, year, semester)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
		return err
	}
	err = c.saveClassInDB(ctx, stuID, year, semester, classes)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to save class[%v %v %v %v]: %v", stuID, year, semester, classes)
		return err
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			err := c.cache.DeleteClassIDList(context.Background(), stuID, year, semester)
			if err != nil {
				classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
			}
		})
	}()
	return nil
}

func (c *ClassRepo) GetClassesFromLocal(ctx context.Context, stuID, year, semester string) ([]*bizmodel.ClassBiz, error) {
	//先从缓存中得到对应的课程id列表
	classIDs, err := c.cache.GetClassIDList(ctx, stuID, year, semester)

	var classes []*model.ClassDO

	//如果获取成功
	if err == nil {
		var waitSearchClassID []string
		//如果classIDs的长度为0,说明key存在,但是无内容，即人为定义的NULL
		//则说明其上一次已经从数据库中查过了，但是也没有获取到
		//我们此时则直接返回nil
		if len(classIDs) == 0 {
			return nil, nil
		}
		tmpClasses, leftID, err1 := c.cache.GetClassesByID(ctx, classIDs...)
		if err1 == nil {
			if len(leftID) == 0 {
				//如果都查询到了，直接返回
				return batchNewBizClasses(tmpClasses), nil
			} else {
				//如果查询成功，但是有部分未查询成功
				//则将leftID添加到等待查询的ID中
				waitSearchClassID = append(waitSearchClassID, leftID...)

				//同时将已经查询到的classes加入到结果集中去
				classes = append(classes, tmpClasses...)
			}
		} else {
			//如果本身查询失败,则将tmpClasses视为无效
			//将所有id都添加到等待查询的ID中
			waitSearchClassID = append(waitSearchClassID, classIDs...)
		}

		//毫无疑问此时,需要从数据库中找到剩余的课程信息
		tmpClasses, err1 = c.getClassesFromDBByIDs(ctx, waitSearchClassID...)
		if err1 != nil {
			return nil, err1
		}

		//将查询到的classes加入到结果集中去
		classes = append(classes, tmpClasses...)

		//另起一个协程来缓存未查询到的课程信息
		go func() {
			err1 = c.cache.AddClass(context.Background(), tmpClasses...)
			if err1 != nil {
				classLog.LogPrinter.Warnf("add class[%v] to cache failed: %v", tmpClasses, err1)
			}
		}()

		//此时可以直接返回
		return batchNewBizClasses(classes), nil
	}

	//如果err!=nil说明不能走缓存了，只能查询数据库
	//并且如果err是cache miss错误，还需要缓存对应的关系
	//课程信息可不缓存，因为可以下次走上面的流程，最终会将课程信息缓存

	var WetherToCache = false

	if errors.Is(err, ErrCacheMiss) {
		WetherToCache = true
	}

	//从数据库中查询数据

	classes, err = c.getClassesFromDB(ctx, stuID, year, semester)
	if err != nil {
		return nil, err
	}
	if WetherToCache {
		//开启协程存取对应关系
		go func() {
			ids := extractClassIDs(classes)
			//如果len(id)==0
			//说明数据库其实没有数据
			//这是我们也得缓存来记录
			//只不过没有id,只有key
			err = c.cache.SetClassIDList(context.Background(), stuID, year, semester, ids...)
		}()
	}

	return batchNewBizClasses(classes), nil
}

func (c *ClassRepo) GetSpecificClassInfo(ctx context.Context, classID string) (*bizmodel.ClassBiz, error) {
	classes, err := c.getClassesFromDBByIDs(ctx, classID)
	if err != nil {
		return nil, err
	}
	if len(classes) == 0 {
		return nil, errors.New("no data")
	}
	return batchNewBizClasses(classes)[0], nil
}

func (c *ClassRepo) UpdateClass(ctx context.Context, stuID, year, semester string, oldClassID string, newClass *bizmodel.ClassBiz) error {
	classDO := model.NewClass(newClass)
	newClass.ID = classDO.ID

	if c.checkSCIdsExist(ctx, stuID, year, semester, classDO.ID) {
		return errcode.ErrClassIsExist
	}

	//先从缓存中删除对应关系
	err := c.cache.DeleteClassIDList(ctx, stuID, year, semester)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
		return err
	}

	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("stu_id = ? AND year = ? AND semester = ? AND cla_id = ?", stuID, year, semester, oldClassID).Delete(&model.StudentClassRelationDO{}).Error
		if err != nil {
			return err
		}
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(classDO).Error
		if err != nil {
			return err
		}
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.StudentClassRelationDO{
			StuID:           stuID,
			ClaID:           classDO.ID,
			Year:            year,
			Semester:        semester,
			IsManuallyAdded: true,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		classLog.LogPrinter.Errorf("update class[%v %v %v oldClassID: %v newclass: %v] failed: %v", stuID, year, semester, oldClassID, newClass)
		return err
	}
	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			err := c.cache.DeleteClassIDList(context.Background(), stuID, year, semester)
			if err != nil {
				classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
			}
		})
	}()

	return nil

}

func (c *ClassRepo) DeleteClass(ctx context.Context, stuID, year, semester string, classID string) error {

	if !c.checkSCIdsExist(ctx, stuID, year, semester, classID) {
		return errcode.ErrSCIDNOTEXIST
	}

	//先从缓存中删除对应关系
	err := c.cache.DeleteClassIDList(ctx, stuID, year, semester)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
		return err
	}

	var isManuallyAdd bool
	err = c.db.WithContext(ctx).Table(model.StudentClassRelationDOTableName).Select("is_manually_added").
		Where("stu_id = ? AND year = ? AND semester = ? AND cla_id = ?", stuID, year, semester, classID).Pluck("is_manually_added", &isManuallyAdd).Error
	if err != nil {
		return err
	}

	err = c.db.WithContext(ctx).Where("stu_id = ? AND year = ? AND semester = ? AND cla_id = ?", stuID, year, semester, classID).Delete(&model.StudentClassRelationDO{}).Error
	if err != nil {
		classLog.LogPrinter.Errorf("delete class[%v %v %v %v] in db failed: %v", stuID, year, semester, classID)
		return err
	}

	err = c.recycleBin.AddClassIDToRecycleBin(ctx, stuID, year, semester, classID, isManuallyAdd)
	if err != nil {
		classLog.LogPrinter.Errorf("failed to add classID[%v %v %v %v] to recycle_bin: %v", stuID, year, semester, classID)
		//不返回
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			err := c.cache.DeleteClassIDList(context.Background(), stuID, year, semester)
			if err != nil {
				classLog.LogPrinter.Errorf("failed to delete class_id_list[%v %v %v] in cache: %v", stuID, year, semester, err)
			}
		})
	}()
	return nil
}

func (c *ClassRepo) saveClassInDB(ctx context.Context, stuID, year, semester string, classes []*bizmodel.ClassBiz) error {
	classesDO := model.BatchNewClasses(classes)
	studentClassRelationsDO := model.BatchNewStudentClassRelationsDO(stuID, year, semester, classes, false)

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("year = ? AND semester = ? AND stu_id = ? AND is_manually_added = false", year, semester, stuID).
			Delete(&model.StudentClassRelationDO{}).Error
		if err != nil {
			return err
		}
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(classesDO).Error
		if err != nil {
			return err
		}
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(studentClassRelationsDO).Error
		if err != nil {
			return err
		}
		return nil
	})

}

func (c *ClassRepo) getClassesFromDB(ctx context.Context, stuID, year, semester string) ([]*model.ClassDO, error) {
	var classes []*model.ClassDO

	err := c.db.WithContext(ctx).Table(model.ClassDOTableName).
		Select("*").
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.id = %s.cla_id", model.StudentClassRelationDOTableName, model.ClassDOTableName, model.StudentClassRelationDOTableName)).
		Where(fmt.Sprintf("%s.stu_id = ? AND %s.year = ? AND %s.semester = ?", model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName),
			stuID, year, semester).
		Find(&classes).Error
	if err != nil {
		return nil, err
	}
	return classes, nil
}

func (c *ClassRepo) getClassesFromDBByIDs(ctx context.Context, classIDs ...string) ([]*model.ClassDO, error) {
	if len(classIDs) == 0 {
		return nil, nil
	}

	var classes []*model.ClassDO
	err := c.db.WithContext(ctx).Table(model.ClassDOTableName).
		Where("id IN ?", classIDs).Find(&classes).Error
	if err != nil {
		return nil, err
	}
	return classes, nil
}

func batchNewBizClasses(classes []*model.ClassDO) []*bizmodel.ClassBiz {
	if len(classes) == 0 {
		return nil
	}

	bizClasses := make([]*bizmodel.ClassBiz, 0, len(classes))
	for _, class := range classes {
		if class == nil {
			continue
		}
		bizClasses = append(bizClasses, &bizmodel.ClassBiz{
			ID:           class.ID,
			Day:          class.Day,
			Teacher:      class.Teacher,
			Where:        class.Where,
			ClassWhen:    class.ClassWhen,
			WeekDuration: class.WeekDuration,
			Classname:    class.Classname,
			Credit:       class.Credit,
			Weeks:        class.Weeks,
			Semester:     class.Semester,
			Year:         class.Year,
			CreatedAt:    class.CreatedAt,
		})
	}
	return bizClasses
}

func extractClassIDs(classes []*model.ClassDO) []string {
	if len(classes) == 0 {
		return nil
	}

	var ids = make([]string, 0, len(classes))

	for _, class := range classes {
		if class == nil {
			continue
		}
		ids = append(ids, class.ID)
	}
	return ids
}
