package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
)

type StudentAndCourseDBRepo struct {
	data *Data
	log  classLog.Clogger
}

func NewStudentAndCourseDBRepo(data *Data, logger classLog.Clogger) *StudentAndCourseDBRepo {
	return &StudentAndCourseDBRepo{
		log:  logger,
		data: data,
	}
}

func (s StudentAndCourseDBRepo) SaveManyStudentAndCourseToDB(ctx context.Context, scs []*model.StudentCourse) error {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)

	// 处理 StudentCourse
	for _, sc := range scs {
		if err := db.Debug().FirstOrCreate(sc).Error; err != nil {
			s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create StudentAndCourses(%v)", sc),
				classLog.Reason, err)
			return errcode.ErrCourseSave
		}
	}
	return nil
}

func (s StudentAndCourseDBRepo) SaveStudentAndCourseToDB(ctx context.Context, sc *model.StudentCourse) error {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.Debug().FirstOrCreate(sc).Error
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create StudentAndCourse(%v)", sc),
			classLog.Reason, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (s StudentAndCourseDBRepo) DeleteStudentAndCourseInDB(ctx context.Context, ID ...string) error {
	if len(ID) == 0 {
		return errors.New("mysql can't delete zero data")
	}
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.Debug().Where("id IN ?", ID).Delete(&model.StudentCourse{}).Error
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:delete in %s where (id = %s)", model.StudentCourseTableName, ID),
			classLog.Reason, err)
		return errcode.ErrClassDelete
	}
	return nil
}
func (s StudentAndCourseDBRepo) CheckExists(ctx context.Context, xnm, xqm, stuId, classId string) bool {
	db := s.data.Mysql.Table(model.StudentCourseTableName).WithContext(ctx)
	var cnt int64
	err := db.Where("stu_id = ? AND cla_id = ? AND year = ? AND semester = ?", stuId, classId, xnm, xqm).Count(&cnt).Error
	if err != nil || cnt == 0 {
		return false
	}
	return true
}
func (s StudentAndCourseDBRepo) CheckIfManuallyAdded(ctx context.Context, classID string) bool {
	db := s.data.Mysql.WithContext(ctx).Table(model.StudentCourseTableName)
	IMA := false
	err := db.Select("is_manually_added").Where("cla_id =?", classID).Limit(1).Scan(&IMA).Error
	if err != nil {
		return false
	}
	return IMA
}

func (s StudentAndCourseDBRepo) GetClassNum(ctx context.Context, stuID, year, semester string, isManuallyAdded bool) (num int64, err error) {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName)
	err = db.Where("stu_id = ? AND year = ? AND semester = ? AND is_manually_added = ?", stuID, year, semester, isManuallyAdded).Count(&num).Error
	if err != nil {
		return 0, err
	}
	return num, nil
}
