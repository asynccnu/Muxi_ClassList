package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClassInfoDBRepo struct {
	data *Data
	log  classLog.Clogger
}

func NewClassInfoDBRepo(data *Data, logger classLog.Clogger) *ClassInfoDBRepo {
	return &ClassInfoDBRepo{
		log:  logger,
		data: data,
	}
}
func (c ClassInfoDBRepo) SaveClassInfosToDB(ctx context.Context, classInfos []*model.ClassInfo) error {
	if len(classInfos) == 0 {
		return nil
	}

	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.Debug().Clauses(clause.OnConflict{DoNothing: true}).Create(&classInfos).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create %v in %s", classInfos, model.ClassInfoTableName),
			classLog.Reason, err)
		return err
	}
	return nil
}

func (c ClassInfoDBRepo) AddClassInfoToDB(ctx context.Context, classInfo *model.ClassInfo) error {
	if classInfo == nil {
		return nil
	}
	db := c.data.DB(ctx).Table(model.ClassInfoTableName).WithContext(ctx)
	err := db.Debug().Clauses(clause.OnConflict{DoNothing: true}).Create(&classInfo).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create %v in %s", classInfo, model.ClassInfoTableName),
			classLog.Reason, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (c ClassInfoDBRepo) GetClassInfoFromDB(ctx context.Context, ID string) (*model.ClassInfo, error) {
	db := c.data.Mysql.Table(model.ClassInfoTableName).WithContext(ctx)
	cla := &model.ClassInfo{}
	err := db.Select([]string{
		"id", "jxb_id", "day", "teacher", "where", "class_when",
		"week_duration", "class_name", "credit", "weeks",
		"semester", "year",
	}).Where("id =?", ID).First(cla).Error
	if err != nil {
		c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfo in %s where (id = %s)", model.ClassInfoTableName, ID),
			classLog.Reason, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrClassNotFound
		}
		return nil, errcode.ErrClassFound
	}
	return cla, err
}

func (c ClassInfoDBRepo) GetClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*model.ClassInfo, error) {
	db := c.data.Mysql.WithContext(ctx)
	var (
		cla = make([]*model.ClassInfo, 0)
	)
	if stuId != "" {
		err := db.Table(model.ClassInfoTableName).
			Select(
				fmt.Sprintf("%s.id", model.ClassInfoTableName),            // 明确指定 class_info 表的 id 列
				fmt.Sprintf("%s.jxb_id", model.ClassInfoTableName),        // 明确指定 class_info 表的 jxb_id 列
				fmt.Sprintf("%s.day", model.ClassInfoTableName),           // 明确指定 class_info 表的 day 列
				fmt.Sprintf("%s.teacher", model.ClassInfoTableName),       // 明确指定 class_info 表的 teacher 列
				fmt.Sprintf("%s.where", model.ClassInfoTableName),         // 明确指定 class_info 表的 where 列
				fmt.Sprintf("%s.class_when", model.ClassInfoTableName),    // 明确指定 class_info 表的 class_when 列
				fmt.Sprintf("%s.week_duration", model.ClassInfoTableName), // 明确指定 class_info 表的 week_duration 列
				fmt.Sprintf("%s.class_name", model.ClassInfoTableName),    // 明确指定 class_info 表的 class_name 列
				fmt.Sprintf("%s.credit", model.ClassInfoTableName),        // 明确指定 class_info 表的 credit 列
				fmt.Sprintf("%s.weeks", model.ClassInfoTableName),         // 明确指定 class_info 表的 weeks 列
				fmt.Sprintf("%s.year", model.ClassInfoTableName),          // 明确指定 class_info 表的 year 列
				fmt.Sprintf("%s.semester", model.ClassInfoTableName),      // 明确指定 class_info 表的 semester 列
			).
			Joins(fmt.Sprintf(
				`LEFT JOIN %s ON %s.id = %s.cla_id`, model.StudentCourseTableName, model.ClassInfoTableName, model.StudentCourseTableName,
			)).
			Where(fmt.Sprintf(
				`%s.stu_id = ? AND %s.year = ? AND %s.semester = ?`, model.StudentCourseTableName, model.StudentCourseTableName, model.StudentCourseTableName),
				stuId, xnm, xqm,
			).
			Order(fmt.Sprintf(
				"%s.day ASC, %s.class_when ASC", model.ClassInfoTableName, model.ClassInfoTableName,
			)).
			Find(&cla).Error
		if err != nil {
			c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfos  where (stu_id = %s,year = %s,semester = %s)",
				stuId, xnm, xqm),
				classLog.Reason, err)
			return nil, err
		}
	} else {
		err := db.Table(model.ClassInfoTableName).
			Select(
				fmt.Sprintf("%s.id", model.ClassInfoTableName),            // 明确指定 class_info 表的 id 列
				fmt.Sprintf("%s.jxb_id", model.ClassInfoTableName),        // 明确指定 class_info 表的 jxb_id 列
				fmt.Sprintf("%s.day", model.ClassInfoTableName),           // 明确指定 class_info 表的 day 列
				fmt.Sprintf("%s.teacher", model.ClassInfoTableName),       // 明确指定 class_info 表的 teacher 列
				fmt.Sprintf("%s.where", model.ClassInfoTableName),         // 明确指定 class_info 表的 where 列
				fmt.Sprintf("%s.class_when", model.ClassInfoTableName),    // 明确指定 class_info 表的 class_when 列
				fmt.Sprintf("%s.week_duration", model.ClassInfoTableName), // 明确指定 class_info 表的 week_duration 列
				fmt.Sprintf("%s.class_name", model.ClassInfoTableName),    // 明确指定 class_info 表的 class_name 列
				fmt.Sprintf("%s.credit", model.ClassInfoTableName),        // 明确指定 class_info 表的 credit 列
				fmt.Sprintf("%s.weeks", model.ClassInfoTableName),         // 明确指定 class_info 表的 weeks 列
				fmt.Sprintf("%s.year", model.ClassInfoTableName),          // 明确指定 class_info 表的 year 列
				fmt.Sprintf("%s.semester", model.ClassInfoTableName),      // 明确指定 class_info 表的 semester 列
			).
			Joins(fmt.Sprintf(
				`LEFT JOIN %s ON %s.id = %s.cla_id`, model.StudentCourseTableName, model.ClassInfoTableName, model.StudentCourseTableName,
			)).
			Where(fmt.Sprintf(
				`%s.year = ? AND %s.semester = ?`, model.StudentCourseTableName, model.StudentCourseTableName),
				xnm, xqm,
			).
			Order(fmt.Sprintf(
				"%s.day ASC, %s.class_when ASC", model.ClassInfoTableName, model.ClassInfoTableName,
			)).
			Find(&cla).Error

		if err != nil {
			c.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classinfos  where (is_manually_added = %v,year = %s,semester = %s)",
				false, xnm, xqm),
				classLog.Reason, err)
			return nil, err
		}
	}
	if len(cla) == 0 {
		c.log.Warnw(classLog.Msg, fmt.Sprintf("Mysql:no class has been found,stuID:%s,year:%s,semester:%s", stuId, xnm, xqm))
		return nil, nil
	}
	return cla, nil
}
