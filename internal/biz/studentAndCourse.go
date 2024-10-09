package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
)

//go:generate mockgen -source=./studentAndCourse.go -destination=./mock/mock_studentAndCourse.go -package=mock_biz
type StudentAndCourseDBRepo interface {
	SaveManyStudentAndCourseToDB(ctx context.Context, scs []*model.StudentCourse) error
	SaveStudentAndCourseToDB(ctx context.Context, sc *model.StudentCourse) error
	GetClassIDsFromSCInDB(ctx context.Context, stuId, xnm, xqm string) ([]string, error)
	DeleteStudentAndCourseInDB(ctx context.Context, ID string) error
	CheckExists(ctx context.Context, xnm, xqm, stuId, classId string) bool
	//GetAllSchoolClassIds(ctx context.Context) ([]string, error)
}

type StudentAndCourseCacheRepo interface {
	SaveManyStudentAndCourseToCache(ctx context.Context, key string, classIds []string) error
	AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error
	GetClassIdsFromCache(ctx context.Context, key string) ([]string, error)
	GetRecycledClassIds(ctx context.Context, key string) ([]string, error)
	DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error
	DeleteAndRecycleClassId(ctx context.Context, deleteKey string, recycleBinKey string, classId string) error
	CheckExists(ctx context.Context, key string, classId string) (bool, error)
	CheckRecycleIdIsExist(ctx context.Context, RecycledBinKey, classId string) bool
	RemoveClassFromRecycledBin(ctx context.Context, RecycledBinKey, classId string) error
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
