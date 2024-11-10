package biz

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
)

type ClassRepo struct {
	ClaRepo *ClassInfoRepo
	Sac     *StudentAndCourseRepo
	TxCtrl  Transaction //控制事务的开启
	log     classLog.Clogger
}

func NewClassRepo(ClaRepo *ClassInfoRepo, TxCtrl Transaction, Sac *StudentAndCourseRepo, logger classLog.Clogger) *ClassRepo {
	return &ClassRepo{
		ClaRepo: ClaRepo,
		Sac:     Sac,
		log:     logger,
		TxCtrl:  TxCtrl,
	}
}

func (cla ClassRepo) SaveClasses(ctx context.Context, r model.SaveClassReq) error {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err1 := cla.ClaRepo.DB.SaveClassInfosToDB(ctx, r.ClassInfos)
		if err1 != nil {
			return fmt.Errorf("error saving class In Transaction: %w", err1)
		}
		err2 := cla.Sac.DB.SaveManyStudentAndCourseToDB(ctx, r.Scs)
		if err2 != nil {
			return fmt.Errorf("error saving studentAndcourse In Transaction: %w", err2)
		}
		return nil
	})
	if errTx != nil {
		cla.log.Errorw(
			classLog.Msg, "func:InTx err",
			classLog.Reason, errTx,
		)
		return errTx
	}

	go func() {
		//缓存
		//如果保存时其value为NULL,则直接覆盖
		err := cla.ClaRepo.Cache.AddClaInfosToCache(context.Background(),
			GenerateClassInfosKey(StuId, Xnm, Xqm),
			r.ClassInfos)
		if err != nil {
			cla.log.Warnw(classLog.Msg, "func:AddClaInfosToCache err",
				classLog.Param, fmt.Sprintf("%v,%v", GenerateClassInfosKey(StuId, Xnm, Xqm), r.ClassInfos),
				classLog.Reason, err)
		}
	}()
	return nil
}

func (cla ClassRepo) GetAllClasses(ctx context.Context) (*model.GetAllClassesResp, error) {
	var (
		cacheGet = true
		StuId    = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm      = model.GetCommonInfoFromCtx(ctx).Year
		Xqm      = model.GetCommonInfoFromCtx(ctx).Semester
		key      = GenerateClassInfosKey(StuId, Xnm, Xqm)
	)

	classInfos, err := cla.ClaRepo.Cache.GetClassInfosFromCache(ctx, key)
	//如果err!=nil(err==redis.Nil)说明该ID第一次进入（redis中没有这个KEY），且未经过数据库，则允许其查数据库，所以要设置cacheGet=false
	//如果err==nil说明其至少经过数据库了，redis中有这个KEY,但可能值为NULL，如果不为NULL，就说明缓存命中了,直接返回没有问题
	//如果为NULL，就说明数据库中没有的数据，其依然在请求，会影响数据库（缓存穿透），我们依然直接返回
	//这时我们就需要直接返回redis中的null，即直接返回nil,而不经过数据库

	if err != nil {
		cacheGet = false
		cla.log.Warnw(classLog.Msg, "func:GetClassInfosFromCache err",
			classLog.Param, fmt.Sprintf("%v", key),
			classLog.Reason, err)
	}
	if !cacheGet {
		//从数据库中获取
		classInfos, err = cla.ClaRepo.DB.GetClassInfos(ctx, StuId, Xnm, Xqm)
		if err != nil {
			cla.log.Errorw(classLog.Msg, "func:GetClassInfos err",
				classLog.Param, fmt.Sprintf("%v,%v,%v", StuId, Xnm, Xqm),
				classLog.Reason, err)
			return nil, errcode.ErrClassNotFound
		}
		go func() {
			//将课程信息当作整体存入redis
			//注意:如果未获取到，即classInfos为nil，redis仍然会设置key-value，只不过value为NULL
			err := cla.ClaRepo.Cache.AddClaInfosToCache(context.Background(), key, classInfos)
			if err != nil {
				cla.log.Warnw(classLog.Msg, "func:AddClaInfosToCache err",
					classLog.Param, fmt.Sprintf("%v,%v", key, classInfos),
					classLog.Reason, err)
			}

		}()
	}
	//检查classInfos是否为空
	//如果不为空，直接返回就好
	//如果为空，则说明没有该数据，需要去查询
	//如果不添加此条件，如果你redis中有值为NULL的话，该值就永远不会更新，所以需要该条件
	//添加该条件，能够让查询数据库的操作效率更高，同时也保证了数据的获取
	if len(classInfos) == 0 {
		return nil, errcode.ErrClassNotFound
	}
	return &model.GetAllClassesResp{ClassInfos: classInfos}, nil
}
func (cla ClassRepo) GetSpecificClassInfo(ctx context.Context, req model.GetSpecificClassInfoReq) (*model.GetSpecificClassInfoResp, error) {
	classInfo, err := cla.ClaRepo.DB.GetClassInfoFromDB(ctx, req.ClassId)
	if err != nil || classInfo == nil {
		cla.log.Errorw(classLog.Msg, "func:GetClassInfoFromDB err",
			classLog.Param, fmt.Sprintf("%v", req.ClassId),
			classLog.Reason, err)
		return nil, errcode.ErrClassNotFound
	}
	return &model.GetSpecificClassInfoResp{ClassInfo: classInfo}, nil
}
func (cla ClassRepo) AddClass(ctx context.Context, req model.AddClassReq) error {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		if err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, req.ClassInfo); err != nil {
			cla.log.Errorw(classLog.Msg, "func:AddClassInfoToDB err",
				classLog.Param, fmt.Sprintf("%v", req.ClassInfo),
				classLog.Reason, err)
			return errcode.ErrClassUpdate
		}

		// 处理 StudentCourse
		if err := cla.Sac.DB.SaveStudentAndCourseToDB(ctx, req.Sc); err != nil {
			cla.log.Errorw(classLog.Msg, "func:SaveStudentAndCourseToDB err",
				classLog.Param, fmt.Sprintf("%v", req.Sc),
				classLog.Reason, err)
			return errcode.ErrClassUpdate
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}
	// 在事务成提交后，异步处理缓存更新
	go func() {
		err := cla.ClaRepo.Cache.UpdateClassInfoInCache(context.Background(), "", GenerateClassInfosKey(StuId, Xnm, Xqm), req.ClassInfo, true)
		if err != nil {
			cla.log.Warnw(classLog.Msg, "func:UpdateClassInfoInCache err",
				classLog.Param, fmt.Sprintf("%v,%v,%v,%v", "", GenerateClassInfosKey(StuId, Xnm, Xqm), req.ClassInfo, true),
				classLog.Reason, err)
		}
	}()
	// 不等待缓存写入完成，直接返回
	return nil
}
func (cla ClassRepo) DeleteClass(ctx context.Context, req model.DeleteClassReq) error {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	//判断该课程是否为手动添加，如果是就同时删除class_info中的数据
	IMA := cla.Sac.DB.CheckIfManuallyAdded(ctx, req.ClassId)
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err := cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, model.GenerateSCID(StuId, req.ClassId, Xnm, Xqm))
		if err != nil {
			return fmt.Errorf("error deleting student: %w", err)
		}
		if IMA {
			err = cla.ClaRepo.DB.DeleteClassInfoInDB(ctx, req.ClassId)
			if err != nil {
				return fmt.Errorf("error deleting classinfo: %w", err)
			}
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}
	key2 := GenerateRecycleSetName(StuId, Xnm, Xqm)
	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, req.ClassId, GenerateClassInfosKey(StuId, Xnm, Xqm))
	if err != nil {
		cla.log.Errorw(classLog.Msg, "func:DeleteClassInfoFromCache err",
			classLog.Param, fmt.Sprintf("%v,%v", req.ClassId, GenerateClassInfosKey(StuId, Xnm, Xqm)),
			classLog.Reason, err)
		return err
	}
	//删除并添加进回收站
	err = cla.Sac.Cache.RecycleClassId(ctx, key2, req.ClassId)
	if err != nil {
		cla.log.Errorw(classLog.Msg, "func:RecycleClassId err",
			classLog.Param, fmt.Sprintf("%v,%v", key2, req.ClassId),
			classLog.Reason, err)
		return err
	}

	return nil
}
func (cla ClassRepo) GetRecycledIds(ctx context.Context) (*model.GetRecycledIdsResp, error) {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	recycleKey := GenerateRecycleSetName(StuId, Xnm, Xqm)
	classIds, err := cla.Sac.Cache.GetRecycledClassIds(ctx, recycleKey)
	if err != nil {
		cla.log.Errorw(classLog.Msg, "func:GetRecycledClassIds err",
			classLog.Param, fmt.Sprintf("%v", recycleKey),
			classLog.Reason, err)
		return nil, err
	}
	return &model.GetRecycledIdsResp{Ids: classIds}, nil
}
func (cla ClassRepo) CheckClassIdIsInRecycledBin(ctx context.Context, req model.CheckClassIdIsInRecycledBinReq) bool {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	RecycledBinKey := GenerateRecycleSetName(StuId, Xnm, Xqm)
	return cla.Sac.Cache.CheckRecycleIdIsExist(ctx, RecycledBinKey, req.ClassId)
}
func (cla ClassRepo) RecoverClassFromRecycledBin(ctx context.Context, req model.RecoverClassFromRecycleBinReq) error {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	RecycledBinKey := GenerateRecycleSetName(StuId, Xnm, Xqm)
	return cla.Sac.Cache.RemoveClassFromRecycledBin(ctx, RecycledBinKey, req.ClassId)
}
func (cla ClassRepo) UpdateClass(ctx context.Context, req model.UpdateClassReq) error {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		//添加新的课程信息
		err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, req.NewClassInfo)
		if err != nil {
			cla.log.Errorw(classLog.Msg, "func:AddClassInfoToDB err",
				classLog.Param, fmt.Sprintf("%v", req.NewClassInfo),
				classLog.Reason, err)
			return errcode.ErrClassUpdate
		}
		//删除原本的学生与课程的对应关系
		err = cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, model.GenerateSCID(StuId, req.OldClassId, Xnm, Xqm))
		if err != nil {
			cla.log.Errorw(classLog.Msg, "func:DeleteStudentAndCourseInDB err",
				classLog.Param, fmt.Sprintf("%v", model.GenerateSCID(StuId, req.OldClassId, Xnm, Xqm)),
				classLog.Reason, err)
			return errcode.ErrClassUpdate
		}
		//添加新的对应关系
		err = cla.Sac.DB.SaveStudentAndCourseToDB(ctx, req.NewSc)
		if err != nil {
			cla.log.Errorw(classLog.Msg, "func:SaveStudentAndCourseToDB err",
				classLog.Param, fmt.Sprintf("%v", req.NewSc),
				classLog.Reason, err)
			return errcode.ErrClassUpdate
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}

	// 缓存相关操作
	go func() {
		//把缓存课表更新
		err := cla.ClaRepo.Cache.UpdateClassInfoInCache(context.Background(),
			req.OldClassId,
			GenerateClassInfosKey(StuId, Xnm, Xqm),
			req.NewClassInfo, false)
		if err != nil {
			cla.log.Warnw(classLog.Msg, "func:UpdateClassInfoInCache err",
				classLog.Param, fmt.Sprintf("%v,%v,%v", req.OldClassId,
					GenerateClassInfosKey(StuId, Xnm, Xqm),
					req.NewClassInfo),
				classLog.Reason, err)
		}
	}()
	return nil
}
func (cla ClassRepo) CheckSCIdsExist(ctx context.Context, req model.CheckSCIdsExistReq) bool {
	var (
		StuId = model.GetCommonInfoFromCtx(ctx).StuId
		Xnm   = model.GetCommonInfoFromCtx(ctx).Year
		Xqm   = model.GetCommonInfoFromCtx(ctx).Semester
	)
	return cla.Sac.DB.CheckExists(ctx, Xnm, Xqm, StuId, req.ClassId)
}
func (cla ClassRepo) GetAllSchoolClassInfos(ctx context.Context) *model.GetAllSchoolClassInfosResp {
	var (
		Xnm = model.GetCommonInfoFromCtx(ctx).Year
		Xqm = model.GetCommonInfoFromCtx(ctx).Semester
	)
	classInfos, err := cla.ClaRepo.DB.GetClassInfos(ctx, "", Xnm, Xqm)
	if err != nil {
		cla.log.Warnw(classLog.Msg, "func:UpdateClassInfoInCache err",
			classLog.Param, fmt.Sprintf("%v,%v", Xnm, Xqm),
			classLog.Reason, err)
		return nil
	}
	return &model.GetAllSchoolClassInfosResp{ClassInfos: classInfos}
}

// ExecuteTransaction 抽象事务的执行函数，接收所有需要在事务内执行的操作
func (cla ClassRepo) executeTransaction(ctx context.Context, operations ...func(ctx context.Context) error) error {
	return cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		for _, operation := range operations {
			if err := operation(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func GenerateRecycleSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("Recycle:%s:%s:%s", stuId, xnm, xqm)
}
func GenerateClassInfosKey(stuId, xnm, xqm string) string {
	return fmt.Sprintf("ClassInfos:%s:%s:%s", stuId, xnm, xqm)
}
