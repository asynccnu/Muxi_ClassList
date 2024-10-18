package data

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
)

type StudentAndCourseDBRepo struct {
	data *Data
	log  *log.Helper
}
type StudentAndCourseCacheRepo struct {
	rdb *redis.Client
	log *log.Helper
}

func NewStudentAndCourseDBRepo(data *Data, logger log.Logger) *StudentAndCourseDBRepo {
	return &StudentAndCourseDBRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}
func NewStudentAndCourseCacheRepo(rdb *redis.Client, logger log.Logger) *StudentAndCourseCacheRepo {
	return &StudentAndCourseCacheRepo{
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

func (s StudentAndCourseCacheRepo) GetRecycledClassIds(ctx context.Context, key string) ([]string, error) {
	res, err := s.rdb.SMembers(key).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:get classIds From set(%s)", key),
			classLog.Reason, err)
		return nil, err
	}
	return res, nil
}
func (s StudentAndCourseCacheRepo) CheckRecycleIdIsExist(ctx context.Context, RecycledBinKey, classId string) bool {
	exists, err := s.rdb.SIsMember(RecycledBinKey, classId).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:check classId(%s) is exist in RecycleBinKey(%s) err", classId, RecycledBinKey),
			classLog.Reason, err)
		return false
	}
	return exists
}
func (s StudentAndCourseCacheRepo) RemoveClassFromRecycledBin(ctx context.Context, RecycledBinKey, classId string) error {
	_, err := s.rdb.SRem(RecycledBinKey, classId).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:remove member(%s) from set(%s) err", classId, RecycledBinKey),
			classLog.Reason, err)
		return err
	}
	return nil
}

func (s StudentAndCourseCacheRepo) RecycleClassId(ctx context.Context, recycleBinKey string, classId string) error {

	// 将 ClassId 放入回收站
	if err := s.rdb.SAdd(recycleBinKey, classId).Err(); err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Add classId(%s) to set(%s) err", classId, recycleBinKey),
			classLog.Reason, err)
		return err
	}

	// 设置回收站的过期时间
	if err := s.rdb.Expire(recycleBinKey, RecycleExpiration).Err(); err != nil {
		s.log.Errorw(classLog.Msg, "Redis:set expire err",
			classLog.Reason, err)
		return err
	}
	return nil
}

func (s StudentAndCourseDBRepo) SaveManyStudentAndCourseToDB(ctx context.Context, scs []*model.StudentCourse) error {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)

	// 处理 StudentCourse
	for _, sc := range scs {
		if err := db.FirstOrCreate(sc).Error; err != nil {
			s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create StudentAndCourses(%v)", sc),
				classLog.Reason, err)
			return errcode.ErrCourseSave
		}
	}
	return nil
}

func (s StudentAndCourseDBRepo) SaveStudentAndCourseToDB(ctx context.Context, sc *model.StudentCourse) error {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.FirstOrCreate(sc).Error
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:create StudentAndCourse(%v)", sc),
			classLog.Reason, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (s StudentAndCourseDBRepo) DeleteStudentAndCourseInDB(ctx context.Context, ID string) error {
	db := s.data.DB(ctx).Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.Where("id =?", ID).Delete(&model.StudentCourse{}).Error
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
