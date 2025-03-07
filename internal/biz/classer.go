package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"time"
)

type ClassUsecase struct {
	classStore          ClassStorage
	recycleBinManager   ClassRecycleBinManager
	manualClassManager  ManualClassManager
	schoolClassExplorer SchoolClassExplorer
	crawler             ClassCrawler
	ccnu                CCNUServiceProxy
	jxbRepo             JxbRepo
}

func NewClassUsecase(classStore ClassStorage,
	recycleBinManager ClassRecycleBinManager,
	manualClassManager ManualClassManager,
	schoolClassExplorer SchoolClassExplorer,
	crawler ClassCrawler,
	JxbRepo JxbRepo,
	Cs CCNUServiceProxy) *ClassUsecase {

	return &ClassUsecase{
		classStore:          classStore,
		recycleBinManager:   recycleBinManager,
		manualClassManager:  manualClassManager,
		schoolClassExplorer: schoolClassExplorer,
		crawler:             crawler,
		jxbRepo:             JxbRepo,
		ccnu:                Cs,
	}
}

func (cluc *ClassUsecase) GetClasses(ctx context.Context, stuID, year, semester string, refresh bool) ([]*model.ClassBiz, error) {
	var (
		classes        = make([]*model.ClassBiz, 0)
		jxbIDs         []string
		SearchFromCCNU = refresh
	)

	if !refresh {
		//直接从数据库中获取课表
		classesFromLocal, err := cluc.classStore.GetClassesFromLocal(ctx, stuID, year, semester)

		if len(classesFromLocal) > 0 {
			classes = classesFromLocal
		}
		if err != nil {
			classLog.LogPrinter.Errorf("get class[%v %v %v] from DB failed: %v", stuID, year, semester, err)
		}

		// 如果数据库中没有
		// 或者时间是每周周一，就(有些特殊时间比如2,9月月末和3,10月月初，默认会优先爬取)默认有0.3的概率去爬取，这样是为了防止课表更新了，但一直会从数据库中获取，导致，课表无法更新
		if err != nil || tool.IsNeedCraw() {
			SearchFromCCNU = true

			crawClasses, jxbids, err := cluc.getCourseFromCrawler(ctx, stuID, year, semester)

			if err != nil {
				classLog.LogPrinter.Errorf("get class[%v %v %v] from CCNU failed: %v", stuID, year, semester, err)
			}

			if err == nil {
				classes = crawClasses
				jxbIDs = jxbids
			}
		}
	} else {
		crawClasses, jxbids, err := cluc.getCourseFromCrawler(ctx, stuID, year, semester)
		if err == nil {
			SearchFromCCNU = true
			classes = crawClasses
			jxbIDs = jxbids

			//从数据库中获取手动添加的课程
			addedClassesFromLocal, err1 := cluc.manualClassManager.GetAddedClasses(ctx, stuID, year, semester)
			if err1 != nil {
				classLog.LogPrinter.Errorf("get added class[%v %v %v] from DB failed: %v", stuID, year, semester, err1)
			}

			if err1 == nil && len(addedClassesFromLocal) > 0 {
				classes = append(classes, addedClassesFromLocal...)
			}
		} else {

			classLog.LogPrinter.Errorf("get class[%v %v %v] from CCNU failed: %v", stuID, year, semester, err)

			//如果爬取失败
			SearchFromCCNU = false

			//使用本地数据库做兜底
			classesFromLocal, err := cluc.classStore.GetClassesFromLocal(ctx, stuID, year, semester)

			if len(classesFromLocal) > 0 {
				classes = classesFromLocal
			}
			if err != nil {
				classLog.LogPrinter.Errorf("get class[%v %v %v] from DB failed: %v", stuID, year, semester, err)
			}
		}
	}

	//如果所有获取途径均失效，则返回错误
	if len(classes) == 0 {
		return nil, errcode.ErrClassNotFound
	}

	if SearchFromCCNU { //如果是从CCNU那边查到的，就存储
		//开个协程来存取
		go func() {
			_ = cluc.classStore.SaveClass(context.Background(), stuID, year, semester, classes)

			//防止ctx因为return就被取消了，所以就改用background，因为这个存取没有精确的要求，所以可以后台完成，用户不需要感知
			_ = cluc.jxbRepo.SaveJxb(context.Background(), stuID, jxbIDs)
		}()
	}
	return classes, nil
}

func (cluc *ClassUsecase) AddClass(ctx context.Context, stuID, year, semester string, info *model.ClassBiz) error {
	//添加课程
	err := cluc.manualClassManager.AddClass(ctx, stuID, year, semester, info)
	if err != nil {
		classLog.LogPrinter.Errorf("Add class[%v %v %v] failed: %v", stuID, year, semester, err)
		return errcode.ErrClassAdd
	}
	return nil
}
func (cluc *ClassUsecase) DeleteClass(ctx context.Context, stuID, year, semester, classId string) error {
	//删除课程
	err := cluc.classStore.DeleteClass(ctx, stuID, year, semester, classId)
	if err != nil {
		classLog.LogPrinter.Errorf("Delete class[%v %v %v] failed: %v", stuID, year, semester, err)
		return errcode.ErrClassDelete
	}
	return nil
}
func (cluc *ClassUsecase) GetRecycledClassInfos(ctx context.Context, stuID, year, semester string) ([]*model.ClassBiz, error) {
	classes, err := cluc.recycleBinManager.GetRecycledClasses(ctx, stuID, year, semester)
	if err != nil {
		classLog.LogPrinter.Errorf("Get recycled class[%v %v %v] failed: %v", stuID, year, semester, err)
		return nil, errcode.ErrGetRecycledClasses
	}
	return classes, nil
}
func (cluc *ClassUsecase) RecoverClassInfo(ctx context.Context, stuID, year, semester, classId string) error {
	//先检查要回复的课程ID是否存在于回收站中
	exist := cluc.recycleBinManager.CheckClassIdIsInRecycledBin(ctx, stuID, year, semester, classId)
	if !exist {
		return errcode.ErrRecycleBinDoNotHaveIt
	}
	err := cluc.recycleBinManager.RecoverClassFromRecycledBin(ctx, stuID, year, semester, classId)
	if err != nil {
		classLog.LogPrinter.Errorf("Recover class[%v %v %v] failed: %v", stuID, year, semester, err)
		return errcode.ErrRecover
	}
	return nil
}
func (cluc *ClassUsecase) UpdateClass(ctx context.Context, stuID, year, semester string, oldClassId string, newClassInfo *model.ClassBiz) error {
	err := cluc.classStore.UpdateClass(ctx, stuID, year, semester, oldClassId, newClassInfo)
	if err != nil {
		classLog.LogPrinter.Errorf("Update class[%v %v %v] failed: %v", stuID, year, semester, err)
		return errcode.ErrClassUpdate
	}
	return nil
}

func (cluc *ClassUsecase) GetAllSchoolClassInfosToOtherService(ctx context.Context, year, semester string, cursor time.Time) []*model.ClassBiz {
	return cluc.schoolClassExplorer.GetAllSchoolClassInfos(ctx, year, semester, cursor)
}

func (cluc *ClassUsecase) GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error) {
	return cluc.jxbRepo.FindStuIdsByJxbId(ctx, jxbId)
}

func (cluc *ClassUsecase) SearchClass(ctx context.Context, classID string) (*model.ClassBiz, error) {
	return cluc.classStore.GetSpecificClassInfo(ctx, classID)
}

func (cluc *ClassUsecase) getCourseFromCrawler(ctx context.Context, stuID string, year string, semester string) ([]*model.ClassBiz, []string, error) {
	////测试用的
	//cookie := "JSESSIONID=77CCA81367438A56D3AFF46797E674A4"

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // 10秒超时,防止影响
	defer cancel()                                                 // 确保在函数返回前取消上下文，防止资源泄漏

	cookie, err := cluc.ccnu.GetCookie(timeoutCtx, stuID)
	if err != nil {
		classLog.LogPrinter.Errorf("Error getting cookie(stu_id:%v) from other service", stuID)
		return nil, nil, err
	}

	var stu Student
	if tool.CheckIsUndergraduate(stuID) { //针对是否是本科生，进行分类
		stu = &Undergraduate{}
	} else {
		stu = &GraduateStudent{}
	}

	return stu.GetClass(ctx, stuID, year, semester, cookie, cluc.crawler)
}
