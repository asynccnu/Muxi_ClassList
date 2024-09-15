package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	log2 "github.com/asynccnu/Muxi_ClassList/internal/logPrinter"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"sync"
)

// ClassCrawler 课程爬虫接口
//
//go:generate mockgen -source=./classer.go -destination=./mock/mock_classer_crawler.go -package=mock_biz
type ClassCrawler interface {
	GetClassInfosForUndergraduate(ctx context.Context, cookie string, xnm, xqm string) ([]*ClassInfo, []*StudentCourse, error)
	GetClassInfoForGraduateStudent(ctx context.Context, cookie string, xnm, xqm string) ([]*ClassInfo, []*StudentCourse, error)
}
type ClassRepoProxy interface {
	SaveClasses(ctx context.Context, stuId, xnm, xqm string, claInfos []*ClassInfo, scs []*StudentCourse) error
	GetAllClasses(ctx context.Context, stuId, xnm, xqm string) ([]*ClassInfo, error)
	GetSpecificClassInfo(ctx context.Context, classId string) (*ClassInfo, error)
	AddClass(ctx context.Context, classInfo *ClassInfo, sc *StudentCourse, xnm, xqm string) error
	DeleteClass(ctx context.Context, classId string, stuId string, xnm string, xqm string) error
	GetRecycledIds(ctx context.Context, stuId, xnm, xqm string) ([]string, error)
	RemoveClassFromRecycledBin(ctx context.Context, stuId, xnm, xqm, classId string) error
	UpdateClass(ctx context.Context, newClassInfo *ClassInfo, newSc *StudentCourse, stuId, oldClassId, xnm, xqm string) error
	CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool
	GetAllSchoolClassInfos(ctx context.Context, xnm, xqm string) []*ClassInfo
	CheckClassIdIsInRecycledBin(ctx context.Context, stuId, xnm, xqm, classId string) bool
}
type ClassUsercase struct {
	ClassRepo ClassRepoProxy
	Crawler   ClassCrawler
	log       log2.LogerPrinter
}

func NewClassUsercase(classRepo ClassRepoProxy, crawler ClassCrawler, log log2.LogerPrinter) *ClassUsercase {
	return &ClassUsercase{
		ClassRepo: classRepo,
		Crawler:   crawler,
		log:       log,
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, StuId string, week int64, xnm, xqm string, cookie string) ([]*Class, error) {
	//var classInfos = make([]*ClassInfo, 0)
	var Scs = make([]*StudentCourse, 0)
	var classes = make([]*Class, 0)
	var err error
	var wg sync.WaitGroup
	classInfos, err := cluc.ClassRepo.GetAllClasses(ctx, StuId, xnm, xqm)
	if err != nil {
		if tool.CheckIsUndergraduate(StuId) { //针对是否是本科生，进行分类
			classInfos, Scs, err = cluc.Crawler.GetClassInfosForUndergraduate(ctx, cookie, xnm, xqm)

			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfosForUndergraduate, err)
				return nil, err
			}
		} else {
			classInfos, Scs, err = cluc.Crawler.GetClassInfoForGraduateStudent(ctx, cookie, xnm, xqm)

			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfoForGraduateStudent, err)
				return nil, err
			}
		}

		err = cluc.ClassRepo.SaveClasses(ctx, StuId, xnm, xqm, classInfos, Scs)
		if err != nil {
			cluc.log.FuncError(cluc.ClassRepo.SaveClasses, err)
			return nil, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cluc.ClassRepo.SaveClasses(ctx, StuId, xnm, xqm, classInfos, Scs)
			if err != nil {
				cluc.log.FuncError(cluc.ClassRepo.SaveClasses, err)
			}
		}()
	}

	for _, classInfo := range classInfos {
		thisWeek := classInfo.SearchWeek(week)
		class := &Class{
			Info:     classInfo,
			ThisWeek: thisWeek && tool.CheckIfThisYear(classInfo.Year, classInfo.Semester),
		}
		classes = append(classes, class)
	}
	wg.Wait()
	return classes, nil
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
func (cluc *ClassUsercase) GetRecycledClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*ClassInfo, error) {
	RecycledClassIds, err := cluc.ClassRepo.GetRecycledIds(ctx, stuId, xnm, xqm)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.GetRecycledIds, err)
		return nil, err
	}
	classInfos := make([]*ClassInfo, 0)
	for _, classId := range RecycledClassIds {
		classInfo, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, classId)
		if err != nil {
			cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
			continue
		}
		classInfos = append(classInfos, classInfo)
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) RecoverClassInfo(ctx context.Context, stuId, xnm, xqm, classId string) error {
	exist := cluc.ClassRepo.CheckClassIdIsInRecycledBin(ctx, stuId, xnm, xqm, classId)
	if !exist {
		return errcode.ErrRecycleBinDoNotHaveIt
	}
	RecycledClassInfo, err := cluc.SearchClass(ctx, classId)
	if err != nil {
		cluc.log.FuncError(cluc.SearchClass, err)
		return errcode.ErrRecover
	}
	err = cluc.AddClass(ctx, stuId, RecycledClassInfo)
	if err != nil {
		cluc.log.FuncError(cluc.AddClass, err)
		return errcode.ErrRecover
	}
	err = cluc.ClassRepo.RemoveClassFromRecycledBin(ctx, stuId, xnm, xqm, classId)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.RemoveClassFromRecycledBin, err)
		return errcode.ErrRecover
	}
	return nil
}
func (cluc *ClassUsercase) SearchClass(ctx context.Context, classId string) (*ClassInfo, error) {
	classInfo, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, classId)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
		return nil, err
	}
	return classInfo, nil
}
func (cluc *ClassUsercase) UpdateClass(ctx context.Context, newClassInfo *ClassInfo, newSc *StudentCourse, stuId, oldClassId, xnm, xqm string) error {
	err := cluc.ClassRepo.UpdateClass(ctx, newClassInfo, newSc, stuId, oldClassId, xnm, xqm)
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.UpdateClass, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool {
	return cluc.ClassRepo.CheckSCIdsExist(ctx, stuId, classId, xnm, xqm)
}
func (cluc *ClassUsercase) GetAllSchoolClassInfosToOtherService(ctx context.Context, xnm, xqm string) []*ClassInfo {
	return cluc.ClassRepo.GetAllSchoolClassInfos(ctx, xnm, xqm)
}
