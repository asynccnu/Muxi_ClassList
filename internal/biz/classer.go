package biz

import (
	"class/internal/errcode"
	log2 "class/internal/logPrinter"
	"context"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type TxController interface {
	Begin(ctx context.Context, db *gorm.DB) *gorm.DB
	RollBack(ctx context.Context, tx *gorm.DB)
	Commit(ctx context.Context, tx *gorm.DB) error
}
type ClassInfoDBRepo interface {
	SaveClassInfosToDB(ctx context.Context, tx *gorm.DB, classInfo []*ClassInfo) error
	AddClassInfoToDB(ctx context.Context, tx *gorm.DB, classInfo *ClassInfo) error
	GetClassInfoFromDB(ctx context.Context, db *gorm.DB, ID string) (*ClassInfo, error)
	DeleteClassInfoInDB(ctx context.Context, tx *gorm.DB, ID string) error
	//UpdateClassInfoInDB(ctx context.Context, tx *gorm.DB, classInfo *ClassInfo) error
}
type ClassInfoCacheRepo interface {
	SaveManyClassInfosToCache(ctx context.Context, keys []string, classInfos []*ClassInfo) error
	AddClassInfoToCache(ctx context.Context, key string, classInfo *ClassInfo) error
	GetClassesFromCache(ctx context.Context, key string) (*ClassInfo, error)
	DeleteClassInfoFromCache(ctx context.Context, key string) error
	UpdateClassInfoInCache(ctx context.Context, key string, classInfo *ClassInfo) error
}
type StudentAndCourseDBRepo interface {
	SaveStudentAndCourseToDB(ctx context.Context, tx *gorm.DB, sc *StudentCourse) error
	SaveManyStudentAndCourseToDB(ctx context.Context, tx *gorm.DB, scs []*StudentCourse) error
	GetClassIDsFromSCInDB(ctx context.Context, db *gorm.DB, stuId, xnm, xqm string) ([]string, error)
	DeleteStudentAndCourseInDB(ctx context.Context, tx *gorm.DB, ID string) error
}
type StudentAndCourseCacheRepo interface {
	SaveManyStudentAndCourseToCache(ctx context.Context, key string, classIds []string) error
	AddStudentAndCourseToCache(ctx context.Context, key string, ClassId string) error
	GetClassIdsFromCache(ctx context.Context, key string) ([]string, error)
	DeleteStudentAndCourseFromCache(ctx context.Context, key string, ClassId string) error
}

// ClassCrawler 课程爬虫接口
type ClassCrawler interface {
	GetClassInfos(ctx context.Context, client *http.Client, xnm, xqm string) ([]*ClassInfo, []*StudentCourse, error)
}
type ClassInfoRepo struct {
	DB    ClassInfoDBRepo
	Cache ClassInfoCacheRepo
}
type StudentAndCourseRepo struct {
	DB    StudentAndCourseDBRepo
	Cache StudentAndCourseCacheRepo
}

func NewClassInfoRepo(DB ClassInfoDBRepo, Cache ClassInfoCacheRepo) *ClassInfoRepo {
	return &ClassInfoRepo{
		DB:    DB,
		Cache: Cache,
	}
}
func NewStudentAndCourseRepo(DB StudentAndCourseDBRepo, Cache StudentAndCourseCacheRepo) *StudentAndCourseRepo {
	return &StudentAndCourseRepo{
		DB:    DB,
		Cache: Cache,
	}
}

type ClassUsercase struct {
	ClassRepo *ClassRepo
	Crawler   ClassCrawler
	log       log2.LogerPrinter
}

func NewClassUsercase(classRepo *ClassRepo, crawler ClassCrawler, log log2.LogerPrinter) *ClassUsercase {
	return &ClassUsercase{
		ClassRepo: classRepo,
		Crawler:   crawler,
		log:       log,
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, StuId string, week int64, xnm, xqm string, client *http.Client) ([]*Class, error) {
	//var classInfos = make([]*ClassInfo, 0)
	var Scs = make([]*StudentCourse, 0)
	var classes = make([]*Class, 0)
	var err error

	classInfos, err := cluc.ClassRepo.GetAllClasses(ctx, StuId, xnm, xqm)
	if err != nil {

		if errors.Is(err, errcode.ErrClassNotFound) {

			classInfos, Scs, err = cluc.Crawler.GetClassInfos(ctx, client, xnm, xqm)

			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfos, err)
				return nil, err
			}

			go func() {
				err := cluc.ClassRepo.SaveClasses(ctx, StuId, xnm, xqm, classInfos, Scs)
				if err != nil {
					cluc.log.FuncError(cluc.ClassRepo.SaveClasses, err)
				}
			}()
		}
		return nil, err
	}

	for _, classInfo := range classInfos {
		thisWeek := classInfo.SearchWeek(week)
		class := &Class{
			Info:     classInfo,
			ThisWeek: thisWeek,
		}
		classes = append(classes, class)
	}

	return classes, nil
}
func (cluc *ClassUsercase) FindClass(ctx context.Context, stuId string, xnm, xqm string, day int64, dur string) ([]*ClassInfo, error) {

	classInfos, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, stuId, xnm, xqm, day, dur)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
		return nil, err
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) AddClass(ctx context.Context, stuId string, info *ClassInfo) error {
	sc := &StudentCourse{
		StuID:           stuId,
		ClaID:           info.ID,
		Year:            info.Year,
		Semester:        info.Semester,
		IsManuallyAdded: true,
	}
	sc.UpdateID()
	err := cluc.ClassRepo.AddClass(ctx, info, sc, info.Year, info.Semester)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.AddClass, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) DeleteClass(ctx context.Context, classId string, stuId string, xnm string, xqm string) error {
	err := cluc.ClassRepo.DeleteClass(ctx, classId, stuId, xnm, xqm)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.DeleteClass, err)
		return err
	}
	return nil
}
