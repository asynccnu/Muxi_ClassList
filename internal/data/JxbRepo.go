package data

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz"
	log "github.com/asynccnu/Muxi_ClassList/internal/logPrinter"
)

type JxbDBRepo struct {
	data *Data
	log  log.LogerPrinter
}

func NewJxbDBRepo(data *Data, log log.LogerPrinter) *JxbDBRepo {
	return &JxbDBRepo{
		data: data,
		log:  log,
	}
}

func (j *JxbDBRepo) SaveJxb(ctx context.Context, jxbId, stuId string) error {
	db := j.data.Mysql.Table(biz.JxbTableName).WithContext(ctx)
	jxb := &biz.Jxb{
		JxbId: jxbId,
		StuId: stuId,
	}
	err := db.FirstOrCreate(jxb).Error
	if err != nil {
		j.log.FuncError(db.Create, err)
		return err
	}
	return nil
}
func (j *JxbDBRepo) FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	var stuIds []string
	err := j.data.Mysql.Raw("SELECT stu_id FROM jxb WHERE jxb_id =  ?", jxbId).Scan(&stuIds).Error
	if err != nil {
		j.log.FuncError(j.data.Mysql.Raw, err)
		return nil, err
	}
	return stuIds, nil
}
