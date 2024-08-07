package data

import (
	"class/internal/biz"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormDatabase struct {
	db  *gorm.DB
	err error
}

func NewGormDatabase(db *gorm.DB) Database {
	return &GormDatabase{
		db: db,
	}
}

func (g *GormDatabase) Begin() Database {
	g.db = g.db.Begin()
	return g
}

func (g *GormDatabase) Create(value interface{}) Database {
	g.err = g.db.Create(value).Clauses(clause.OnConflict{ //如果主键冲突，忽略冲突
		DoNothing: true,
	}).Error
	return g
}

func (g *GormDatabase) WithContext(ctx context.Context) Database {
	g.db = g.db.WithContext(ctx)
	return g
}

func (g *GormDatabase) Table(name string) Database {
	g.db = g.db.Table(name)
	return g
}

func (g *GormDatabase) Commit() error {
	return g.db.Commit().Error
}

func (g *GormDatabase) Rollback() {
	g.db.Rollback()
}

func (g *GormDatabase) Error() error {
	return g.err
}
func (g *GormDatabase) GetClassInfos(ctx context.Context, claId, xnm, xqm string) ([]*biz.ClassInfo, error) {
	classInfos := make([]*biz.ClassInfo, 0)
	db := g.db.Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.Where("class_id = ? AND year = ? AND semester = ?", claId, xnm, xqm).Find(&classInfos).Error
	return classInfos, err
}
func (g *GormDatabase) GetSpecificClassInfos(ctx context.Context, Id string) (*biz.ClassInfo, error) {
	classInfo := &biz.ClassInfo{}
	db := g.db.Table(biz.ClassInfoTableName).WithContext(ctx)
	err := db.Where("id = ?", Id).First(&classInfo).Error
	return classInfo, err
}
func (g *GormDatabase) DeleteClassInfo(ctx context.Context, id string) error {
	err := g.db.Where("id =?", id).Delete(&biz.ClassInfo{}).Error
	return err
}
func (g *GormDatabase) GetClassIds(ctx context.Context, stuId string) ([]string, error) {
	var classIds []string
	db := g.db.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.Where("stu_id = ?", stuId).
		Select("cla_id").
		Pluck("cla_id", &classIds).Error
	return classIds, err
}
