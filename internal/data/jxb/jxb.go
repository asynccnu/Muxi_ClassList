package jxb

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/data/jxb/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JxbDBRepo struct {
	db *gorm.DB
}

func NewJxbDBRepo(db *gorm.DB) *JxbDBRepo {
	return &JxbDBRepo{
		db: db,
	}
}

func (j *JxbDBRepo) SaveJxb(ctx context.Context, stuID string, jxbID []string) error {
	if len(jxbID) == 0 {
		return nil
	}

	db := j.db.Table(model.JxbTableName).WithContext(ctx)
	var jxb = make([]model.Jxb, 0, len(jxbID))
	for _, id := range jxbID {
		jxb = append(jxb, model.Jxb{
			JxbId: id,
			StuId: stuID,
		})
	}

	err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&jxb).Error
	if err != nil {
		//TODO:log
		return err
	}
	return nil
}
func (j *JxbDBRepo) FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	var stuIds []string
	err := j.db.Table(model.JxbTableName).Select("stu_id").Where("jxb_id = ?", jxbId).Pluck("stu_id", &stuIds).Error
	if err != nil {
		return nil, err
	}
	return stuIds, nil
}
