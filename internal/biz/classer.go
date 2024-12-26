package biz

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	model2 "github.com/asynccnu/Muxi_ClassList/internal/model"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"sync"
	"time"
)

type ClassUsercase struct {
	classRepo ClassRepoProxy
	crawler   ClassCrawler
	ccnu      CCNUServiceProxy
	jxbRepo   JxbRepo
	log       classLog.Clogger
}

func NewClassUsercase(classRepo ClassRepoProxy, crawler ClassCrawler, JxbRepo JxbRepo, Cs CCNUServiceProxy, logger classLog.Clogger) *ClassUsercase {
	return &ClassUsercase{
		classRepo: classRepo,
		crawler:   crawler,
		jxbRepo:   JxbRepo,
		ccnu:      Cs,
		log:       logger,
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, stuID, year, semester string, week int64) ([]*model2.Class, error) {
	var (
		scs            = make([]*model2.StudentCourse, 0)
		classes        = make([]*model2.Class, 0)
		classInfos     = make([]*model2.ClassInfo, 0)
		wg             sync.WaitGroup
		SearchFromCCNU = false
	)
	resp1, err := cluc.classRepo.GetAllClasses(ctx, model2.GetAllClassesReq{
		StuID:    stuID,
		Year:     year,
		Semester: semester,
	})
	if resp1 != nil {
		classInfos = resp1.ClassInfos
	}
	// 如果数据库中没有
	// 或者时间是每周周一，就(有些特殊时间比如2,9月月末和3,10月月初，默认会优先爬取)默认有0.3的概率去爬取，这样是为了防止课表更新了，但一直会从数据库中获取，导致，课表无法更新
	if err != nil || tool.IsNeedCraw() {
		SearchFromCCNU = true
		////测试用的
		//cookie := "JSESSIONID=B3414E736467BF833BAA58CF866974A3"

		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond) // 1秒超时,防止影响
		defer cancel()                                                        // 确保在函数返回前取消上下文，防止资源泄漏

		cookie, err := cluc.ccnu.GetCookie(timeoutCtx, stuID)
		if err != nil {
			//封装class
			wc := model2.WrapClassInfo(classInfos)
			classes, _ = wc.ConvertToClass(week)
			return classes, nil
		}

		var stu Student
		if tool.CheckIsUndergraduate(stuID) { //针对是否是本科生，进行分类
			stu = &Undergraduate{}
		} else {
			stu = &GraduateStudent{}
		}
		classInfos, scs, err = stu.GetClass(ctx, stuID, year, semester, cookie, cluc.crawler)
		if err != nil {
			return nil, err
		}

		//存课程
		wg.Add(1)
		go func() {
			defer wg.Done()
			cluc.classRepo.CheckAndSaveClass(context.Background(), stuID, year, semester, classInfos, scs)

		}()
	}
	//封装class
	wc := model2.WrapClassInfo(classInfos)
	classes, jxbIDs := wc.ConvertToClass(week)
	wg.Wait()
	if SearchFromCCNU { //如果是从CCNU那边查到的，就存下jxb_id
		//开个协程来存取jxb
		go func() {
			//防止ctx因为return就被取消了，所以就改用background，因为这个存取没有精确的要求，所以可以后台完成，用户不需要感知
			if err := cluc.jxbRepo.SaveJxb(context.Background(), stuID, jxbIDs); err != nil {
				cluc.log.Warnw(classLog.Msg, "func:SaveClasses err",
					classLog.Param, fmt.Sprintf("%v,%v", stuID, jxbIDs),
					classLog.Reason, err)
			}

		}()
	}
	return classes, nil
}

func (cluc *ClassUsercase) AddClass(ctx context.Context, stuID string, info *model2.ClassInfo) error {
	sc := &model2.StudentCourse{
		StuID:           stuID,
		ClaID:           info.ID,
		Year:            info.Year,
		Semester:        info.Semester,
		IsManuallyAdded: true,
	}
	if cluc.classRepo.CheckSCIdsExist(ctx, model2.CheckSCIdsExistReq{StuID: stuID, Year: info.Year, Semester: info.Semester, ClassId: info.ID}) {
		return errcode.ErrClassIsExist
	}
	err := cluc.classRepo.AddClass(ctx, model2.AddClassReq{
		StuID:     stuID,
		Year:      info.Year,
		Semester:  info.Semester,
		ClassInfo: info,
		Sc:        sc,
	})
	if err != nil {
		return err
	}
	return nil
}
func (cluc *ClassUsercase) DeleteClass(ctx context.Context, stuID, year, semester, classId string) error {
	err := cluc.classRepo.DeleteClass(ctx, model2.DeleteClassReq{
		StuID:    stuID,
		Year:     year,
		Semester: semester,
		ClassId:  []string{classId},
	})
	if err != nil {
		return errcode.ErrClassDelete
	}
	return nil
}
func (cluc *ClassUsercase) GetRecycledClassInfos(ctx context.Context, stuID, year, semester string) ([]*model2.ClassInfo, error) {
	RecycledClassIds, err := cluc.classRepo.GetRecycledIds(ctx, model2.GetRecycledIdsReq{
		StuID:    stuID,
		Year:     year,
		Semester: semester,
	})
	if err != nil {
		return nil, err
	}
	classInfos := make([]*model2.ClassInfo, 0)
	for _, classId := range RecycledClassIds.Ids {
		resp, err := cluc.classRepo.GetSpecificClassInfo(ctx, model2.GetSpecificClassInfoReq{
			StuID:    stuID,
			Year:     year,
			Semester: semester,
			ClassId:  classId})
		if err != nil {
			continue
		}
		classInfos = append(classInfos, resp.ClassInfo)
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) RecoverClassInfo(ctx context.Context, stuID, year, semester, classId string) error {
	exist := cluc.classRepo.CheckClassIdIsInRecycledBin(ctx, model2.CheckClassIdIsInRecycledBinReq{
		StuID:    stuID,
		Year:     year,
		Semester: semester,
		ClassId:  classId,
	})
	if !exist {
		return errcode.ErrRecycleBinDoNotHaveIt
	}
	RecycledClassInfo, err := cluc.SearchClass(ctx, classId)
	if err != nil {
		return errcode.ErrRecover
	}
	err = cluc.AddClass(ctx, stuID, RecycledClassInfo)
	if err != nil {
		return errcode.ErrRecover
	}
	err = cluc.classRepo.RecoverClassFromRecycledBin(ctx, model2.RecoverClassFromRecycleBinReq{
		ClassId: classId,
	})
	if err != nil {
		return errcode.ErrRecover
	}
	return nil
}
func (cluc *ClassUsercase) SearchClass(ctx context.Context, classId string) (*model2.ClassInfo, error) {
	resp, err := cluc.classRepo.GetSpecificClassInfo(ctx, model2.GetSpecificClassInfoReq{ClassId: classId})
	if err != nil {
		return nil, err
	}
	return resp.ClassInfo, nil
}
func (cluc *ClassUsercase) UpdateClass(ctx context.Context, stuID, year, semester string, newClassInfo *model2.ClassInfo, newSc *model2.StudentCourse, oldClassId string) error {
	err := cluc.classRepo.UpdateClass(ctx, model2.UpdateClassReq{
		StuID:        stuID,
		Year:         year,
		Semester:     semester,
		NewClassInfo: newClassInfo,
		NewSc:        newSc,
		OldClassId:   oldClassId,
	})
	if err != nil {
		return err
	}
	return nil
}
func (cluc *ClassUsercase) CheckSCIdsExist(ctx context.Context, stuID, year, semester, classId string) bool {
	return cluc.classRepo.CheckSCIdsExist(ctx, model2.CheckSCIdsExistReq{
		StuID:    stuID,
		Year:     year,
		Semester: semester,
		ClassId:  classId,
	})
}
func (cluc *ClassUsercase) GetAllSchoolClassInfosToOtherService(ctx context.Context, year, semester string) []*model2.ClassInfo {
	return cluc.classRepo.GetAllSchoolClassInfos(ctx, model2.GetAllSchoolClassInfosReq{
		Year:     year,
		Semester: semester,
	}).ClassInfos
}
func (cluc *ClassUsercase) GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	return cluc.jxbRepo.FindStuIdsByJxbId(ctx, jxbId)
}
