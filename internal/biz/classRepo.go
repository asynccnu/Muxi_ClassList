package biz

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	model2 "github.com/asynccnu/Muxi_ClassList/internal/model"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

const MaxNum = 20

type ClassRepo struct {
	ClaRepo *ClassInfoRepo
	Sac     *StudentAndCourseRepo
	TxCtrl  Transaction //控制事务的开启
	log     *log.Helper
}

func NewClassRepo(ClaRepo *ClassInfoRepo, TxCtrl Transaction, Sac *StudentAndCourseRepo, logger log.Logger) *ClassRepo {
	return &ClassRepo{
		ClaRepo: ClaRepo,
		Sac:     Sac,
		log:     log.NewHelper(logger),
		TxCtrl:  TxCtrl,
	}
}

func (cla ClassRepo) GetAllClasses(ctx context.Context, req model2.GetAllClassesReq) (*model2.GetAllClassesResp, error) {
	var (
		cacheGet = true
		key      = GenerateClassInfosKey(req.StuID, req.Year, req.Semester)
	)

	classInfos, err := cla.ClaRepo.Cache.GetClassInfosFromCache(ctx, key)
	//如果err!=nil(err==redis.Nil)说明该ID第一次进入（redis中没有这个KEY），且未经过数据库，则允许其查数据库，所以要设置cacheGet=false
	//如果err==nil说明其至少经过数据库了，redis中有这个KEY,但可能值为NULL，如果不为NULL，就说明缓存命中了,直接返回没有问题
	//如果为NULL，就说明数据库中没有的数据，其依然在请求，会影响数据库（缓存穿透），我们依然直接返回
	//这时我们就需要直接返回redis中的null，即直接返回nil,而不经过数据库

	if err != nil {
		cacheGet = false
		cla.log.Warnf("Get Class [%+v] From Cache failed: %v", req, err)
	}
	if !cacheGet {
		//从数据库中获取
		classInfos, err = cla.ClaRepo.DB.GetClassInfos(ctx, req.StuID, req.Year, req.Semester)
		if err != nil {
			cla.log.Errorf("Get Class [%+v] From DB failed: %v", req, err)
			return nil, errcode.ErrClassFound
		}
		go func() {
			//将课程信息当作整体存入redis
			//注意:如果未获取到，即classInfos为nil，redis仍然会设置key-value，只不过value为NULL
			_ = cla.ClaRepo.Cache.AddClaInfosToCache(context.Background(), key, classInfos)
		}()
	}
	//检查classInfos是否为空
	//如果不为空，直接返回就好
	//如果为空，则说明没有该数据，需要去查询
	//如果不添加此条件，即便你redis中有值为NULL的话，也不会返回错误，就导致不会去爬取更新，所以需要该条件
	//添加该条件，能够让查询数据库的操作效率更高，同时也保证了数据的获取
	if len(classInfos) == 0 {
		return nil, errcode.ErrClassNotFound
	}
	return &model2.GetAllClassesResp{ClassInfos: classInfos}, nil
}
func (cla ClassRepo) GetSpecificClassInfo(ctx context.Context, req model2.GetSpecificClassInfoReq) (*model2.GetSpecificClassInfoResp, error) {
	classInfo, err := cla.ClaRepo.DB.GetClassInfoFromDB(ctx, req.ClassId)
	if err != nil || classInfo == nil {
		return nil, errcode.ErrClassNotFound
	}
	return &model2.GetSpecificClassInfoResp{ClassInfo: classInfo}, nil
}
func (cla ClassRepo) AddClass(ctx context.Context, req model2.AddClassReq) error {
	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, GenerateClassInfosKey(req.StuID, req.Year, req.Semester))
	if err != nil {
		return err
	}
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		if err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, req.ClassInfo); err != nil {
			return errcode.ErrClassUpdate
		}
		// 处理 StudentCourse
		if err := cla.Sac.DB.SaveStudentAndCourseToDB(ctx, req.Sc); err != nil {
			return errcode.ErrClassUpdate
		}
		cnt, err := cla.Sac.DB.GetClassNum(ctx, req.StuID, req.Year, req.Semester, req.Sc.IsManuallyAdded)
		if err == nil && cnt > MaxNum {
			return fmt.Errorf("class num limit")
		}
		return nil
	})
	if errTx != nil {
		cla.log.Errorf("Add Class [%+v] failed:%v", req, errTx)
		return errTx
	}
	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			_ = cla.ClaRepo.Cache.DeleteClassInfoFromCache(context.Background(), GenerateClassInfosKey(req.StuID, req.Year, req.Semester))
		})
	}()
	return nil
}
func (cla ClassRepo) DeleteClass(ctx context.Context, req model2.DeleteClassReq) error {

	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, GenerateClassInfosKey(req.StuID, req.Year, req.Semester))
	if err != nil {
		cla.log.Errorf("Delete Class [%+v] from Cache failed:%v", req, err)
		return err
	}
	//删除并添加进回收站
	recycleSetName := GenerateRecycleSetName(req.StuID, req.Year, req.Semester)
	err = cla.Sac.Cache.RecycleClassId(ctx, recycleSetName, req.ClassId...)
	if err != nil {
		cla.log.Errorf("Add Class [%+v] To RecycleBin failed:%v", req, err)
		return err
	}
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err := cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, req.StuID, req.Year, req.Semester, req.ClassId)
		if err != nil {
			return fmt.Errorf("error deleting student: %w", err)
		}
		return nil
	})
	if errTx != nil {
		cla.log.Errorf("Delete Class [%+v] In DB failed:%v", req, errTx)
		return errTx
	}
	return nil
}
func (cla ClassRepo) GetRecycledIds(ctx context.Context, req model2.GetRecycledIdsReq) (*model2.GetRecycledIdsResp, error) {
	recycleKey := GenerateRecycleSetName(req.StuID, req.Year, req.Semester)
	classIds, err := cla.Sac.Cache.GetRecycledClassIds(ctx, recycleKey)
	if err != nil {
		return nil, err
	}
	return &model2.GetRecycledIdsResp{Ids: classIds}, nil
}
func (cla ClassRepo) CheckClassIdIsInRecycledBin(ctx context.Context, req model2.CheckClassIdIsInRecycledBinReq) bool {

	RecycledBinKey := GenerateRecycleSetName(req.StuID, req.Year, req.Semester)
	return cla.Sac.Cache.CheckRecycleIdIsExist(ctx, RecycledBinKey, req.ClassId)
}
func (cla ClassRepo) RecoverClassFromRecycledBin(ctx context.Context, req model2.RecoverClassFromRecycleBinReq) error {
	RecycledBinKey := GenerateRecycleSetName(req.StuID, req.Year, req.Semester)
	return cla.Sac.Cache.RemoveClassFromRecycledBin(ctx, RecycledBinKey, req.ClassId)
}
func (cla ClassRepo) UpdateClass(ctx context.Context, req model2.UpdateClassReq) error {
	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, GenerateClassInfosKey(req.StuID, req.Year, req.Semester))
	if err != nil {
		return err
	}
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		//添加新的课程信息
		err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, req.NewClassInfo)
		if err != nil {
			return errcode.ErrClassUpdate
		}
		//删除原本的学生与课程的对应关系
		err = cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, req.StuID, req.Year, req.Semester, []string{req.OldClassId})
		if err != nil {
			return errcode.ErrClassUpdate
		}
		//添加新的对应关系
		err = cla.Sac.DB.SaveStudentAndCourseToDB(ctx, req.NewSc)
		if err != nil {
			return errcode.ErrClassUpdate
		}
		return nil
	})
	if errTx != nil {
		cla.log.Errorf("Update Class [%+v] In DB  failed:%v", req, errTx)
		return errTx
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			_ = cla.ClaRepo.Cache.DeleteClassInfoFromCache(context.Background(), GenerateClassInfosKey(req.StuID, req.Year, req.Semester))
		})
	}()

	return nil
}

// 检查下原来的课程和要添加的课程是否一致
// 并做出相应变化
func (cla ClassRepo) SaveClass(ctx context.Context, stuID, year, semester string, classInfos []*model2.ClassInfo, scs []*model2.StudentCourse) {
	key := GenerateClassInfosKey(stuID, year, semester)

	_ = cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, key)

	err := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err := cla.Sac.DB.DeleteStudentAndCourseByTimeFromDB(ctx, stuID, year, semester)
		if err != nil {
			return err
		}
		err = cla.ClaRepo.DB.SaveClassInfosToDB(ctx, classInfos)
		if err != nil {
			return err
		}
		err = cla.Sac.DB.SaveManyStudentAndCourseToDB(ctx, scs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		cla.log.Errorw(classLog.Msg, fmt.Sprintf("save class[%v\n%v] in db err:%v", classInfos, scs, err))
	}

	go func() {
		//延迟双删
		time.AfterFunc(1*time.Second, func() {
			_ = cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, key)
		})
	}()
}

func (cla ClassRepo) CheckSCIdsExist(ctx context.Context, req model2.CheckSCIdsExistReq) bool {
	return cla.Sac.DB.CheckExists(ctx, req.Year, req.Semester, req.StuID, req.ClassId)
}
func (cla ClassRepo) GetAllSchoolClassInfos(ctx context.Context, req model2.GetAllSchoolClassInfosReq) *model2.GetAllSchoolClassInfosResp {

	classInfos, err := cla.ClaRepo.DB.GetClassInfos(ctx, "", req.Year, req.Semester)
	if err != nil {
		return nil
	}
	return &model2.GetAllSchoolClassInfosResp{ClassInfos: classInfos}
}

func GenerateRecycleSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("Recycle:%s:%s:%s", stuId, xnm, xqm)
}
func GenerateClassInfosKey(stuId, xnm, xqm string) string {
	return fmt.Sprintf("ClassInfos:%s:%s:%s", stuId, xnm, xqm)
}
