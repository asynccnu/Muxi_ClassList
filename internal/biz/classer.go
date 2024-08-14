package biz

import (
	"class/internal/errcode"
	log2 "class/internal/logPrinter"
	"context"
	"errors"
	"net/http"
)

// ClassCrawler 课程爬虫接口
type ClassCrawler interface {
	GetClassInfos(ctx context.Context, client *http.Client, xnm, xqm string) ([]*ClassInfo, []*StudentCourse, error)
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

//func (cluc *ClassUsercase) FindClass(ctx context.Context, stuId string, xnm, xqm string, day int64, dur string) ([]*ClassInfo, error) {
//
//	classInfos, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, stuId, xnm, xqm, day, dur)
//	if err != nil {
//		cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
//		return nil, err
//	}
//	return classInfos, nil
//}

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
