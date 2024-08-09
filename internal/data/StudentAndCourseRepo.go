package data

import (
	"class/internal/biz"
	"class/internal/errcode"
	log "class/internal/logPrinter"
	"context"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type studentAndCourseDBRepo struct {
	log log.LogerPrinter
}
type studentAndCourseCacheRepo struct {
	rdb *redis.Client
	log log.LogerPrinter
}

func NewStudentAndCourseDBRepo(log log.LogerPrinter) biz.StudentAndCourseDBRepo {
	return &studentAndCourseDBRepo{
		log: log,
	}
}
func NewStudentAndCourseCacheRepo(rdb *redis.Client, log log.LogerPrinter) biz.StudentAndCourseCacheRepo {
	return &studentAndCourseCacheRepo{
		rdb: rdb,
		log: log,
	}
}
func (s studentAndCourseCacheRepo) SaveManyStudentAndCourseToCache(ctx context.Context, key string, classIds []string) error {
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
func (s studentAndCourseCacheRepo) AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error {

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

func (s studentAndCourseCacheRepo) GetClassIdsFromCache(ctx context.Context, key string) ([]string, error) {
	res, err := s.rdb.SMembers(key).Result()
	if err != nil {
		s.log.FuncError(s.rdb.SMembers, err)
		return nil, err
	}
	return res, nil
}

func (s studentAndCourseCacheRepo) DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error {
	_, err := s.rdb.SRem(key, ClassId).Result()
	if err != nil {
		s.log.FuncError(s.rdb.SRem, err)
		return err
	}
	return nil
}

func (s studentAndCourseDBRepo) SaveManyStudentAndCourseToDB(ctx context.Context, tx *gorm.DB, scs []*biz.StudentCourse) error {
	tx = tx.Table(biz.StudentCourseTableName).WithContext(ctx)

	// 处理 StudentCourse
	for _, sc := range scs {
		if err := tx.Create(sc).Clauses(clause.OnConflict{ //如果主键冲突，忽略冲突
			DoNothing: true,
		}).Error; err != nil {
			s.log.FuncError(tx.Create, err)
			return errcode.ErrCourseSave
		}
	}
	return nil
}

func (s studentAndCourseDBRepo) SaveStudentAndCourseToDB(ctx context.Context, tx *gorm.DB, sc *biz.StudentCourse) error {
	tx = tx.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := tx.Create(sc).Clauses(clause.OnConflict{ //如果主键冲突，忽略冲突
		DoNothing: true,
	}).Error
	if err != nil {
		s.log.FuncError(tx.Create, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (s studentAndCourseDBRepo) GetClassIDsFromSCInDB(ctx context.Context, db *gorm.DB, stuId, xnm, xqm string) ([]string, error) {
	var classIds []string
	db = db.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := db.Where("stu_id = ? AND year = ? AND semester", stuId, xnm, xqm).
		Select("cla_id").
		Pluck("cla_id", &classIds).Error
	return classIds, err
}

func (s studentAndCourseDBRepo) DeleteStudentAndCourseInDB(ctx context.Context, tx *gorm.DB, ID string) error {
	tx = tx.Table(biz.StudentCourseTableName).WithContext(ctx)
	err := tx.Where("id =?", ID).Delete(&biz.StudentCourse{}).Error
	if err != nil {
		s.log.FuncError(tx.Delete, err)
		return errcode.ErrClassDelete
	}
	return nil
}
