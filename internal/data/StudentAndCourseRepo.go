package data

import (
	"context"
	"errors"
	"github.com/asynccnu/Muxi_ClassList/internal/biz"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	log "github.com/asynccnu/Muxi_ClassList/internal/logPrinter"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type StudentAndCourseDBRepo struct {
	data *Data
	log  log.LogerPrinter
}
type StudentAndCourseCacheRepo struct {
	rdb *redis.Client
	log log.LogerPrinter
}

func NewStudentAndCourseDBRepo(data *Data, log log.LogerPrinter) *StudentAndCourseDBRepo {
	return &StudentAndCourseDBRepo{
		log:  log,
		data: data,
	}
}
func NewStudentAndCourseCacheRepo(rdb *redis.Client, log log.LogerPrinter) *StudentAndCourseCacheRepo {
	return &StudentAndCourseCacheRepo{
		rdb: rdb,
		log: log,
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
					s.log.FuncError(pipe.SAdd, err)
					return err
				}

			}
			return nil
		})
		return err
	}, key) // 监控所有要操作的 keys
	if err != nil {
		s.log.FuncError(s.rdb.Watch, err)
		return err
	}
	// 设置 Set 的过期时间
	err = s.rdb.Expire(key, Expiration).Err()
	if err != nil {
		s.log.FuncError(s.rdb.Expire, err)
		return err
	}

	return nil
}
func (s StudentAndCourseCacheRepo) AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error {

	err := s.rdb.SAdd(key, ClassId).Err()
	if err != nil {
		s.log.FuncError(s.rdb.SAdd, err)
		return err
	}
	// 设置 Set 的过期时间
	err = s.rdb.Expire(key, Expiration).Err()
	if err != nil {
		s.log.FuncError(s.rdb.Expire, err)
		return err
	}
	return nil
}

func (s StudentAndCourseCacheRepo) GetClassIdsFromCache(ctx context.Context, key string) ([]string, error) {
	res, err := s.rdb.SMembers(key).Result()
	if err != nil {
		s.log.FuncError(s.rdb.SMembers, err)
		return nil, err
	}
	return res, nil
}

func (s StudentAndCourseCacheRepo) DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error {
	_, err := s.rdb.SRem(key, ClassId).Result()
	if err != nil {
		s.log.FuncError(s.rdb.SRem, err)
		return err
	}
	return nil
}
func (s StudentAndCourseCacheRepo) CheckExists(ctx context.Context, key string, classId string) (bool, error) {
	exists, err := s.rdb.SIsMember(key, classId).Result()
	if err != nil {
		s.log.FuncError(s.rdb.SIsMember, err)
		return false, err
	}
	return exists, nil
}
func (s StudentAndCourseDBRepo) SaveManyStudentAndCourseToDB(ctx context.Context, scs []*biz.StudentCourse) error {
	db := s.data.DB(ctx).Table(biz.StudentCourseTableName).WithContext(ctx)

	// 处理 StudentCourse
	for _, sc := range scs {
		if err := db.FirstOrCreate(sc).Error; err != nil {
			s.log.FuncError(db.Create, err)
			return errcode.ErrCourseSave
		}
	}
	return nil
}

func (s StudentAndCourseDBRepo) SaveStudentAndCourseToDB(ctx context.Context, sc *biz.StudentCourse) error {
	db := s.data.DB(ctx).Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.FirstOrCreate(sc).Error
	if err != nil {
		s.log.FuncError(db.Create, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (s StudentAndCourseDBRepo) GetClassIDsFromSCInDB(ctx context.Context, stuId, xnm, xqm string) ([]string, error) {
	var classIds []string
	db := s.data.Mysql.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.Where("stu_id = ? AND year = ? AND semester", stuId, xnm, xqm).
		Select("cla_id").
		Pluck("cla_id", &classIds).Error
	return classIds, err
}

func (s StudentAndCourseDBRepo) DeleteStudentAndCourseInDB(ctx context.Context, ID string) error {
	db := s.data.DB(ctx).Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.Where("id =?", ID).Delete(&biz.StudentCourse{}).Error
	if err != nil {
		s.log.FuncError(db.Delete, err)
		return errcode.ErrClassDelete
	}
	return nil
}
func (s StudentAndCourseDBRepo) CheckExists(ctx context.Context, xnm, xqm, stuId, classId string) bool {
	db := s.data.Mysql.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.Where("stu_id = ? AND cla_id = ? AND year = ? AND semester", stuId, classId, xnm, xqm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else {
		return true
	}
}
