package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
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
func (s StudentAndCourseCacheRepo) SaveManyStudentAndCourseToCache(ctx context.Context, key string, classIds []string) error {
	err := s.rdb.Watch(func(tx *redis.Tx) error {
		// 开始事务
		_, err := tx.TxPipelined(func(pipe redis.Pipeliner) error {
			for _, classId := range classIds {
				// 将 classId 添加到对应的 Set 中
				err := pipe.SAdd(key, classId).Err()
				if err != nil {
					s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Add classId(%s) to set(%s) err", classId, key),
						classLog.Reason, err)
					return err
				}

			}
			return nil
		})
		return err
	}, key) // 监控所有要操作的 keys
	if err != nil {
		s.log.Errorw(classLog.Msg, "Redis:watch SaveManyStudentAndCourseToCache err",
			classLog.Reason, err)
		return err
	}
	// 设置 Set 的过期时间
	err = s.rdb.Expire(key, Expiration).Err()
	if err != nil {
		s.log.Errorw(classLog.Msg, "Redis:set expire err",
			classLog.Reason, err)
		return err
	}

	return nil
}
func (s StudentAndCourseCacheRepo) AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error {

	err := s.rdb.SAdd(key, ClassId).Err()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Add classId(%s) to set(%s) err", ClassId, key),
			classLog.Reason, err)
		return err
	}
	// 设置 Set 的过期时间
	err = s.rdb.Expire(key, Expiration).Err()
	if err != nil {
		s.log.Errorw(classLog.Msg, "Redis:set expire err",
			classLog.Reason, err)
		return err
	}
	return nil
}

func (s StudentAndCourseCacheRepo) GetClassIdsFromCache(ctx context.Context, key string) ([]string, error) {
	res, err := s.rdb.SMembers(key).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:get classIds From set(%s)", key),
			classLog.Reason, err)
		return nil, err
	}
	return res, nil
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
func (s StudentAndCourseCacheRepo) DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error {
	_, err := s.rdb.SRem(key, ClassId).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:remove member(%s) from set(%s) err", ClassId, key),
			classLog.Reason, err)
		return err
	}
	return nil
}

func (s StudentAndCourseCacheRepo) DeleteAndRecycleClassId(ctx context.Context, deleteKey string, recycleBinKey string, classId string) error {
	// 开启事务
	_, err := s.rdb.TxPipelined(func(pipe redis.Pipeliner) error {
		// 删除 ClassId
		if err := pipe.SRem(deleteKey, classId).Err(); err != nil {
			s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:remove member(%s) from set(%s) err", classId, deleteKey),
				classLog.Reason, err)
			return err
		}

		// 将 ClassId 放入回收站
		if err := pipe.SAdd(recycleBinKey, classId).Err(); err != nil {
			s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:Add classId(%s) to set(%s) err", classId, recycleBinKey),
				classLog.Reason, err)
			return err
		}

		// 设置回收站的过期时间
		if err := pipe.Expire(recycleBinKey, RecycleExpiration).Err(); err != nil {
			s.log.Errorw(classLog.Msg, "Redis:set expire err",
				classLog.Reason, err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
func (s StudentAndCourseCacheRepo) CheckExists(ctx context.Context, key string, classId string) (bool, error) {
	exists, err := s.rdb.SIsMember(key, classId).Result()
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Redis:check classId(%s) is exist in set(%s) err", classId, key),
			classLog.Reason, err)
		return false, err
	}
	return exists, nil
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

func (s StudentAndCourseDBRepo) GetClassIDsFromSCInDB(ctx context.Context, stuId, xnm, xqm string) ([]string, error) {
	var classIds []string
	db := s.data.Mysql.Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.Where("stu_id = ? AND year = ? AND semester = ?", stuId, xnm, xqm).
		Select("cla_id").
		Pluck("cla_id", &classIds).Error
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find classids from Table %s where (stu_id = %s,year = %s,semester = %s)",
			model.StudentCourseTableName, stuId, xnm, xqm),
			classLog.Reason, err)
		return nil, err
	}
	return classIds, nil
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
	err := db.Where("stu_id = ? AND cla_id = ? AND year = ? AND semester = ?", stuId, classId, xnm, xqm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else {
		return true
	}
}
func (s StudentAndCourseDBRepo) GetAllSchoolClassIds(ctx context.Context) ([]string, error) {
	var classIds []string
	db := s.data.Mysql.Table(model.StudentCourseTableName).WithContext(ctx)
	err := db.Where("is_manually_added = ?", false).
		Select("cla_id").
		Pluck("cla_id", &classIds).Error
	if err != nil {
		s.log.Errorw(classLog.Msg, fmt.Sprintf("Mysql:find all classids from Table %s where (is_manually_added = false)", model.StudentCourseTableName),
			classLog.Reason, err)
		return nil, err
	}
	return classIds, nil
}
