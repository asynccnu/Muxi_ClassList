package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
)

//go:generate mockgen -source=./studentAndCourse.go -destination=./mock/mock_studentAndCourse.go -package=mock_biz
type StudentAndCourseDBRepo interface {
	SaveManyStudentAndCourseToDB(ctx context.Context, scs []*model.StudentCourse) error
	SaveStudentAndCourseToDB(ctx context.Context, sc *model.StudentCourse) error
	DeleteStudentAndCourseInDB(ctx context.Context, ID string) error
	CheckExists(ctx context.Context, xnm, xqm, stuId, classId string) bool
}

type StudentAndCourseCacheRepo interface {
	GetRecycledClassIds(ctx context.Context, key string) ([]string, error)
	RecycleClassId(ctx context.Context, recycleBinKey string, classId string) error
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
