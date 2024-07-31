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
func (g *GormDatabase) GetClassInfos(id, xnm, xqm string) ([]*biz.ClassInfo, error) {
	classInfos := make([]*biz.ClassInfo, 0)
	err := g.db.Where("stu_id = ? AND year = ? AND semester = ?", id, xnm, xqm).Find(&classInfos).Error
	return classInfos, err
}
func (g *GormDatabase) GetSpecificClassInfos(id string, xnm, xqm string, day int64, dur string) ([]*biz.ClassInfo, error) {
	classInfos := make([]*biz.ClassInfo, 0)
	err := g.db.Where("stu_id = ? AND year = ?  AND semester = ? AND day = ? AND class_when = ? ", id, xnm, xqm, day, dur).Find(&classInfos).Error
	return classInfos, err
}
func (g *GormDatabase) DeleteClassInfo(id string) error {
	err := g.db.Where("id =?", id).Delete(&biz.ClassInfo{}).Error
	return err
}
