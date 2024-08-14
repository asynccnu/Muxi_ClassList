package biz

import "context"

type StudentAndCourseDBRepo interface {
	SaveManyStudentAndCourseToDB(ctx context.Context, scs []*StudentCourse) error
	SaveStudentAndCourseToDB(ctx context.Context, sc *StudentCourse) error
	GetClassIDsFromSCInDB(ctx context.Context, stuId, xnm, xqm string) ([]string, error)
	DeleteStudentAndCourseInDB(ctx context.Context, ID string) error
}
type StudentAndCourseCacheRepo interface {
	SaveManyStudentAndCourseToCache(ctx context.Context, key string, classIds []string) error
	AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error
	GetClassIdsFromCache(ctx context.Context, key string) ([]string, error)
	DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error
}

type StudentAndCourseRepo struct {
	DB    StudentAndCourseDBRepo
	Cache StudentAndCourseCacheRepo
}

func NewStudentAndCourseRepo(DB StudentAndCourseDBRepo, Cache StudentAndCourseCacheRepo) *StudentAndCourseRepo {
	return &StudentAndCourseRepo{
		DB:    DB,
		Cache: Cache,
	}
}
