package data

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
)

type JxbDBRepo struct {
	data *Data
	log  classLog.Clogger
}

func NewJxbDBRepo(data *Data, logger classLog.Clogger) *JxbDBRepo {
	return &JxbDBRepo{
		data: data,
		log:  logger,
	}
}

func (j *JxbDBRepo) SaveJxb(ctx context.Context, jxbId, stuId string) error {
	db := j.data.Mysql.Table(model.JxbTableName).WithContext(ctx)
	jxb := &model.Jxb{
		JxbId: jxbId,
		StuId: stuId,
	}
	err := db.Where("jxb_id = ? AND stu_id = ?", jxb.JxbId, jxb.StuId).FirstOrCreate(jxb).Error
	if err != nil {
		j.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:Save Jxb{jxb_id = %s, stu_id = %s} err)", jxbId, stuId),
			classLog.Reason, err)
		return err
	}
	return nil
}
func (j *JxbDBRepo) FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	var stuIds []string
	err := j.data.Mysql.Raw("SELECT stu_id FROM jxb WHERE jxb_id =  ?", jxbId).Scan(&stuIds).Error
	if err != nil {
		j.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:Find StuIds By JxbId(%s) err", jxbId),
			classLog.Reason, err)
		return nil, err
	}
	return stuIds, nil
}
