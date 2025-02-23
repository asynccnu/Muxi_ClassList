package repo

import (
	"context"
	bizmodel "github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ClassRepo struct {
	cache ClassCache
	db    *gorm.DB
}

func (c *ClassRepo) SaveClass(ctx context.Context, stuID, year, semester string, classes []bizmodel.ClassBiz) error {
	err := c.cache.DeleteClassIDList(ctx, stuID, year, semester)
	if err != nil {
		//TODO:log
		return err
	}
	err = c.saveClassInDB(ctx, stuID, year, semester, classes)
	if err != nil {
		//TODO:log
		//cla.log.Errorw(classLog.Msg, fmt.Sprintf("save class[%v %v] in db err:%v", classInfos, scs, err))
		return err
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			_ = cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, key)
		})
	}()
}

func (c *ClassRepo) GetClassesFromLocal(ctx context.Context, stuID, year, semester string) ([]bizmodel.ClassBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ClassRepo) GetSpecificClassInfo(ctx context.Context, classID string) (bizmodel.ClassBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ClassRepo) UpdateClass(ctx context.Context, stuID, year, semester string, oldClassID string, newClass bizmodel.ClassBiz) error {
	//TODO implement me
	panic("implement me")
}

func (c *ClassRepo) DeleteClass(ctx context.Context, stuID, year, semester string, classID string) error {
	//TODO implement me
	panic("implement me")
}

func (c *ClassRepo) saveClassInDB(ctx context.Context, stuID, year, semester string, classes []bizmodel.ClassBiz) error {
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
