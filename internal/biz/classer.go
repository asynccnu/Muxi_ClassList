package biz

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"sync"
	"time"
)

type ClassCrawler interface {
	GetClassInfosForUndergraduate(ctx context.Context, req model.GetClassInfosForUndergraduateReq) (*model.GetClassInfosForUndergraduateResp, error)
	GetClassInfoForGraduateStudent(ctx context.Context, req model.GetClassInfoForGraduateStudentReq) (*model.GetClassInfoForGraduateStudentResp, error)
}
type ClassRepoProxy interface {
	SaveClasses(ctx context.Context, req model.SaveClassReq) error
	GetAllClasses(ctx context.Context) (*model.GetAllClassesResp, error)
	GetSpecificClassInfo(ctx context.Context, req model.GetSpecificClassInfoReq) (*model.GetSpecificClassInfoResp, error)
	AddClass(ctx context.Context, req model.AddClassReq) error
	DeleteClass(ctx context.Context, req model.DeleteClassReq) error
	GetRecycledIds(ctx context.Context) (*model.GetRecycledIdsResp, error)
	RecoverClassFromRecycledBin(ctx context.Context, req model.RecoverClassFromRecycleBinReq) error
	UpdateClass(ctx context.Context, req model.UpdateClassReq) error
	CheckSCIdsExist(ctx context.Context, req model.CheckSCIdsExistReq) bool
	GetAllSchoolClassInfos(ctx context.Context) *model.GetAllSchoolClassInfosResp
	CheckClassIdIsInRecycledBin(ctx context.Context, req model.CheckClassIdIsInRecycledBinReq) bool
}
type JxbRepo interface {
	SaveJxb(ctx context.Context, jxbId, stuId string) error
	FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
type CCNUServiceProxy interface {
	GetCookie(ctx context.Context) (string, error)
}
type ClassUsercase struct {
	ClassRepo ClassRepoProxy
	Crawler   ClassCrawler
	Cs        CCNUServiceProxy
	JxbRepo   JxbRepo
	log       *log.Helper
}

func NewClassUsercase(classRepo ClassRepoProxy, crawler ClassCrawler, JxbRepo JxbRepo, Cs CCNUServiceProxy, logger log.Logger) *ClassUsercase {
	return &ClassUsercase{
		ClassRepo: classRepo,
		Crawler:   crawler,
		JxbRepo:   JxbRepo,
		Cs:        Cs,
		log:       log.NewHelper(logger),
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, week int64) ([]*model.Class, error) {
	var (
		Scs            = make([]*model.StudentCourse, 0)
		Jxbmp          = make(map[string]struct{}, 10)
		classes        = make([]*model.Class, 0)
		classInfos     = make([]*model.ClassInfo, 0)
		wg             sync.WaitGroup
		SearchFromCCNU = false
		StuId          = model.GetCommonInfoFromCtx(ctx).StuId
	)
	resp1, err := cluc.ClassRepo.GetAllClasses(ctx)
	if resp1 != nil {
		classInfos = resp1.ClassInfos
	}
	// 如果数据库中没有
	// 或者时间是每周周一，就(有些特殊时间比如2,9月月末和3,10月月初，默认会优先爬取)默认有0.3的概率去爬取，这样是为了防止课表更新了，但一直会从数据库中获取，导致，课表无法更新
	if err != nil || tool.IsNeedCraw() {
		SearchFromCCNU = true
		////测试用的
		//cookie := "JSESSIONID=E6F1CDB285CE1833B6C07B7EEACD6255"

		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond) // 1秒超时,防止影响
		defer cancel()                                                        // 确保在函数返回前取消上下文，防止资源泄漏

		cookie, err := cluc.Cs.GetCookie(timeoutCtx)
		if err != nil {
			cluc.log.Warnw(classLog.Msg, "func:GetCookie err",
				classLog.Reason, err)
			//封装class
			for _, classInfo := range classInfos {
				thisWeek := classInfo.SearchWeek(week)
				class := &model.Class{
					Info:     classInfo,
					ThisWeek: thisWeek && tool.CheckIfThisYear(classInfo.Year, classInfo.Semester),
				}
				classes = append(classes, class)
			}
			return classes, err
		}
		if tool.CheckIsUndergraduate(StuId) { //针对是否是本科生，进行分类

			resp, err := cluc.Crawler.GetClassInfosForUndergraduate(ctx, model.GetClassInfosForUndergraduateReq{
				Cookie: cookie,
			})
			if resp.ClassInfos != nil {
				classInfos = resp.ClassInfos
			}
			if resp.StudentCourses != nil {
				Scs = resp.StudentCourses
			}
			if err != nil {
				return nil, err
			}
		} else {
			resp2, err := cluc.Crawler.GetClassInfoForGraduateStudent(ctx, model.GetClassInfoForGraduateStudentReq{
				Cookie: cookie,
			})
			if resp2.ClassInfos != nil {
				classInfos = resp2.ClassInfos
			}
			if resp2.StudentCourses != nil {
				Scs = resp2.StudentCourses
			}
			if err != nil {
				return nil, err
			}
		}
		//存课程
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cluc.ClassRepo.SaveClasses(ctx, model.SaveClassReq{
				ClassInfos: classInfos,
				Scs:        Scs,
			})
			if err != nil {
				cluc.log.Warnw(classLog.Msg, "func:SaveClasses err",
					classLog.Param, fmt.Sprintf("%v", model.SaveClassReq{
						ClassInfos: classInfos,
						Scs:        Scs,
					}),
					classLog.Reason, err)
			}
		}()
	}

	for _, classInfo := range classInfos {
		thisWeek := classInfo.SearchWeek(week)
		class := &model.Class{
			Info:     classInfo,
			ThisWeek: thisWeek && tool.CheckIfThisYear(classInfo.Year, classInfo.Semester),
		}
		if class.Info.JxbId != "" {
			Jxbmp[class.Info.JxbId] = struct{}{}
		}
		classes = append(classes, class)
	}
	wg.Wait()
	if SearchFromCCNU { //如果是从CCNU那边查到的，就存下jxb_id
		//开个协程来存取jxb
		go func() {
			var err error
			for k := range Jxbmp {
				//防止ctx因为return就被取消了，所以就改用background，因为这个存取没有精确的要求，所以可以后台完成，用户不需要感知
				err = cluc.JxbRepo.SaveJxb(context.Background(), k, StuId)
				if err != nil {
					cluc.log.Warnw(classLog.Msg, "func:SaveClasses err",
						classLog.Param, fmt.Sprintf("%v,%v", k, StuId),
						classLog.Reason, err)
				}
			}
		}()
	}
	return classes, nil
}

func (cluc *ClassUsercase) AddClass(ctx context.Context, info *model.ClassInfo) error {
	var (
		stuId = model.GetCommonInfoFromCtx(ctx).StuId
	)
	sc := &model.StudentCourse{
		StuID:           stuId,
		ClaID:           info.ID,
		Year:            info.Year,
		Semester:        info.Semester,
		IsManuallyAdded: true,
	}
	sc.UpdateID()
	if cluc.ClassRepo.CheckSCIdsExist(ctx, model.CheckSCIdsExistReq{ClassId: info.ID}) {
		cluc.log.Warnw(classLog.Msg, fmt.Sprintf("the id(%s) is try add class(%s) which is existed", stuId, info.ID),
			classLog.Reason, errcode.ErrClassIsExist)
		return errcode.ErrClassIsExist
	}
	err := cluc.ClassRepo.AddClass(ctx, model.AddClassReq{
		ClassInfo: info,
		Sc:        sc,
	})
	if err != nil {
		return err
	}
	return nil
}
func (cluc *ClassUsercase) DeleteClass(ctx context.Context, classId string) error {
	err := cluc.ClassRepo.DeleteClass(ctx, model.DeleteClassReq{
		ClassId: classId,
	})
	if err != nil {
		return err
	}
	return nil
}
func (cluc *ClassUsercase) GetRecycledClassInfos(ctx context.Context) ([]*model.ClassInfo, error) {
	RecycledClassIds, err := cluc.ClassRepo.GetRecycledIds(ctx)
	if err != nil {
		return nil, err
	}
	classInfos := make([]*model.ClassInfo, 0)
	for _, classId := range RecycledClassIds.Ids {
		resp, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, model.GetSpecificClassInfoReq{ClassId: classId})
		if err != nil {
			continue
		}
		classInfos = append(classInfos, resp.ClassInfo)
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) RecoverClassInfo(ctx context.Context, classId string) error {
	exist := cluc.ClassRepo.CheckClassIdIsInRecycledBin(ctx, model.CheckClassIdIsInRecycledBinReq{
		ClassId: classId,
	})
	if !exist {
		return errcode.ErrRecycleBinDoNotHaveIt
	}
	RecycledClassInfo, err := cluc.SearchClass(ctx, classId)
	if err != nil {
		return errcode.ErrRecover
	}
	err = cluc.AddClass(ctx, RecycledClassInfo)
	if err != nil {
		return errcode.ErrRecover
	}
	err = cluc.ClassRepo.RecoverClassFromRecycledBin(ctx, model.RecoverClassFromRecycleBinReq{
		ClassId: classId,
	})
	if err != nil {
		return errcode.ErrRecover
	}
	return nil
}
func (cluc *ClassUsercase) SearchClass(ctx context.Context, classId string) (*model.ClassInfo, error) {
	resp, err := cluc.ClassRepo.GetSpecificClassInfo(ctx, model.GetSpecificClassInfoReq{ClassId: classId})
	if err != nil {
		return nil, err
	}
	return resp.ClassInfo, nil
}
func (cluc *ClassUsercase) UpdateClass(ctx context.Context, newClassInfo *model.ClassInfo, newSc *model.StudentCourse, oldClassId string) error {
	err := cluc.ClassRepo.UpdateClass(ctx, model.UpdateClassReq{
		NewClassInfo: newClassInfo,
		NewSc:        newSc,
		OldClassId:   oldClassId,
	})
	if err != nil {
		return err
	}
	return nil
}
func (cluc *ClassUsercase) CheckSCIdsExist(ctx context.Context, classId string) bool {
	return cluc.ClassRepo.CheckSCIdsExist(ctx, model.CheckSCIdsExistReq{
		ClassId: classId,
	})
}
func (cluc *ClassUsercase) GetAllSchoolClassInfosToOtherService(ctx context.Context) []*model.ClassInfo {
	return cluc.ClassRepo.GetAllSchoolClassInfos(ctx).ClassInfos
}
func (cluc *ClassUsercase) GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	return cluc.JxbRepo.FindStuIdsByJxbId(ctx, jxbId)
}
