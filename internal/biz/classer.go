package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	log2 "github.com/asynccnu/Muxi_ClassList/internal/logPrinter"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"sync"
)

// ClassCrawler 课程爬虫接口
//
//go:generate mockgen -source=./classer.go -destination=./mock/mock_classer_crawler.go -package=mock_biz
type ClassCrawler interface {
	GetClassInfosForUndergraduate(ctx context.Context, req model.GetClassInfosForUndergraduateReq) (*model.GetClassInfosForUndergraduateResp, error)
	GetClassInfoForGraduateStudent(ctx context.Context, req model.GetClassInfoForGraduateStudentReq) (*model.GetClassInfoForGraduateStudentResp, error)
}
type ClassRepoProxy interface {
	SaveClasses(ctx context.Context, req model.SaveClassReq) error
	GetAllClasses(ctx context.Context, req model.GetAllClassesReq) (*model.GetAllClassesResp, error)
	GetSpecificClassInfo(ctx context.Context, req model.GetSpecificClassInfoReq) (*model.GetSpecificClassInfoResp, error)
	AddClass(ctx context.Context, req model.AddClassReq) error
	DeleteClass(ctx context.Context, req model.DeleteClassReq) error
	GetRecycledIds(ctx context.Context, req model.GetRecycledIdsReq) (*model.GetRecycledIdsResp, error)
	RecoverClassFromRecycledBin(ctx context.Context, req model.RecoverClassFromRecycleBinReq) error
	UpdateClass(ctx context.Context, req model.UpdateClassReq) error
	CheckSCIdsExist(ctx context.Context, req model.CheckSCIdsExistReq) bool
	GetAllSchoolClassInfos(ctx context.Context, req model.GetAllSchoolClassInfosReq) *model.GetAllSchoolClassInfosResp
	CheckClassIdIsInRecycledBin(ctx context.Context, req model.CheckClassIdIsInRecycledBinReq) bool
}
type JxbRepo interface {
	SaveJxb(ctx context.Context, jxbId, stuId string) error
	FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
type ClassUsercase struct {
	ClassRepo ClassRepoProxy
	Crawler   ClassCrawler
	JxbRepo   JxbRepo
	log       log2.LogerPrinter
}

func NewClassUsercase(classRepo ClassRepoProxy, crawler ClassCrawler, JxbRepo JxbRepo, log log2.LogerPrinter) *ClassUsercase {
	return &ClassUsercase{
		ClassRepo: classRepo,
		Crawler:   crawler,
		JxbRepo:   JxbRepo,
		log:       log,
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, StuId string, week int64, xnm, xqm string, cookie string) ([]*model.Class, error) {
	//var classInfos = make([]*ClassInfo, 0)
	var Scs = make([]*model.StudentCourse, 0)
	var Jxbmp = make(map[string]struct{}, 10)
	var classes = make([]*model.Class, 0)
	var classInfos = make([]*model.ClassInfo, 0)
	var wg sync.WaitGroup
	resp1, err := cluc.ClassRepo.GetAllClasses(ctx, model.GetAllClassesReq{
		StuId: StuId,
		Xnm:   xnm,
		Xqm:   xqm,
	})
	if resp1 != nil {
		classInfos = resp1.ClassInfos
	}
	// 如果数据库中没有
	// 或者时间是每周周一，就(有些特殊时间比如2,9月月末和3,10月月初，默认会优先爬取)默认有0.3的概率去爬取，这样是为了防止课表更新了，但一直会从数据库中获取，导致，课表无法更新
	if err != nil || tool.IsNeedCraw() {
		if tool.CheckIsUndergraduate(StuId) { //针对是否是本科生，进行分类
			resp, err := cluc.Crawler.GetClassInfosForUndergraduate(ctx, model.GetClassInfosForUndergraduateReq{
				Cookie: cookie,
				Xnm:    xnm,
				Xqm:    xqm,
			})
			if resp.ClassInfos != nil {
				classInfos = resp.ClassInfos
			}
			if resp.StudentCourses != nil {
				Scs = resp.StudentCourses
			}
			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfosForUndergraduate, err)
				return nil, err
			}
		} else {
			resp2, err := cluc.Crawler.GetClassInfoForGraduateStudent(ctx, model.GetClassInfoForGraduateStudentReq{
				Cookie: cookie,
				Xnm:    xnm,
				Xqm:    xqm,
			})
			if resp2.ClassInfos != nil {
				classInfos = resp2.ClassInfos
			}
			if resp2.StudentCourses != nil {
				Scs = resp2.StudentCourses
			}
			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfoForGraduateStudent, err)
				return nil, err
			}
		}
		//存课程
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cluc.ClassRepo.SaveClasses(ctx, model.SaveClassReq{
				StuId:      StuId,
				Xnm:        xnm,
				Xqm:        xqm,
				ClassInfos: classInfos,
				Scs:        Scs,
			})
			if err != nil {
				cluc.log.FuncError(cluc.ClassRepo.SaveClasses, err)
			}
		}()
	}

	for _, classInfo := range classInfos {
		thisWeek := classInfo.SearchWeek(week)
		class := &model.Class{
			Info:     classInfo,
			ThisWeek: thisWeek && tool.CheckIfThisYear(classInfo.Year, classInfo.Semester),
		}
		Jxbmp[class.Info.JxbId] = struct{}{}
		classes = append(classes, class)
	}
	wg.Wait()
	//开个协程来存取jxb
	go func() {
		var err error
		for k, _ := range Jxbmp {
			//防止ctx因为return就被取消了，所以就改用background，因为这个存取没有精确的要求，所以可以后台完成，用户不需要感知
			err = cluc.JxbRepo.SaveJxb(context.Background(), k, StuId)
			if err != nil {
				cluc.log.FuncError(cluc.JxbRepo.SaveJxb, err)
			}
		}
	}()
	return classes, nil
}

func (cluc *ClassUsercase) AddClass(ctx context.Context, stuId string, info *model.ClassInfo) error {
	sc := &model.StudentCourse{
		StuID:           stuId,
		ClaID:           info.ID,
		Year:            info.Year,
		Semester:        info.Semester,
		IsManuallyAdded: true,
	}
	sc.UpdateID()
	err := cluc.ClassRepo.AddClass(ctx, model.AddClassReq{
		ClassInfo: info,
		Sc:        sc,
		Xnm:       info.Year,
		Xqm:       info.Semester,
	})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.AddClass, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) DeleteClass(ctx context.Context, classId string, stuId string, xnm string, xqm string) error {
	err := cluc.ClassRepo.DeleteClass(ctx, model.DeleteClassReq{
		ClassId: classId,
		StuId:   stuId,
		Xnm:     xnm,
		Xqm:     xqm,
	})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.DeleteClass, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) GetRecycledClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*model.ClassInfo, error) {
	RecycledClassIds, err := cluc.ClassRepo.GetRecycledIds(ctx, model.GetRecycledIdsReq{
		StuId: stuId,
		Xnm:   xnm,
		Xqm:   xqm,
	})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.GetRecycledIds, err)
		return nil, err
	}
	classInfos := make([]*model.ClassInfo, 0)
	for _, classId := range RecycledClassIds.Ids {
		resp, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, model.GetSpecificClassInfoReq{ClassId: classId})
		if err != nil {
			cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
			continue
		}
		classInfos = append(classInfos, resp.ClassInfo)
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) RecoverClassInfo(ctx context.Context, stuId, xnm, xqm, classId string) error {
	exist := cluc.ClassRepo.CheckClassIdIsInRecycledBin(ctx, model.CheckClassIdIsInRecycledBinReq{
		StuId:   stuId,
		Xnm:     xnm,
		Xqm:     xqm,
		ClassId: classId,
	})
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
	err = cluc.ClassRepo.RecoverClassFromRecycledBin(ctx, model.RecoverClassFromRecycleBinReq{
		StuId:   stuId,
		Xnm:     xnm,
		Xqm:     xqm,
		ClassId: classId,
	})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.RecoverClassFromRecycledBin, err)
		return errcode.ErrRecover
	}
	return nil
}
func (cluc *ClassUsercase) SearchClass(ctx context.Context, classId string) (*model.ClassInfo, error) {
	resp, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, model.GetSpecificClassInfoReq{ClassId: classId})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.GetSpecificClassInfo, err)
		return nil, err
	}
	return resp.ClassInfo, nil
}
func (cluc *ClassUsercase) UpdateClass(ctx context.Context, newClassInfo *model.ClassInfo, newSc *model.StudentCourse, stuId, oldClassId, xnm, xqm string) error {
	err := cluc.ClassRepo.UpdateClass(ctx, model.UpdateClassReq{
		NewClassInfo: newClassInfo,
		NewSc:        newSc,
		StuId:        stuId,
		OldClassId:   oldClassId,
		Xnm:          xnm,
		Xqm:          xqm,
	})
	if err != nil {
		cluc.log.FuncError(cluc.ClassRepo.UpdateClass, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool {
	return cluc.ClassRepo.CheckSCIdsExist(ctx, model.CheckSCIdsExistReq{
		StuId:   stuId,
		ClassId: classId,
		Xnm:     xnm,
		Xqm:     xqm,
	})
}
func (cluc *ClassUsercase) GetAllSchoolClassInfosToOtherService(ctx context.Context, xnm, xqm string) []*model.ClassInfo {
	return cluc.ClassRepo.GetAllSchoolClassInfos(ctx, model.GetAllSchoolClassInfosReq{
		Xnm: xnm,
		Xqm: xqm,
	}).ClassInfos
}
func (cluc *ClassUsercase) GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	return cluc.JxbRepo.FindStuIdsByJxbId(ctx, jxbId)
}
