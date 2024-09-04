package biz

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	log "github.com/asynccnu/Muxi_ClassList/internal/logPrinter"
)

type ClassRepo struct {
	ClaRepo *ClassInfoRepo
	Sac     *StudentAndCourseRepo
	TxCtrl  Transaction //控制事务的开启
	log     log.LogerPrinter
}

func NewClassRepo(ClaRepo *ClassInfoRepo, TxCtrl Transaction, Sac *StudentAndCourseRepo, log log.LogerPrinter) *ClassRepo {
	return &ClassRepo{
		ClaRepo: ClaRepo,
		Sac:     Sac,
		log:     log,
		TxCtrl:  TxCtrl,
	}
}

func (cla ClassRepo) SaveClasses(ctx context.Context, stuId, xnm, xqm string, claInfos []*ClassInfo, scs []*StudentCourse) error {
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err1 := cla.ClaRepo.DB.SaveClassInfosToDB(ctx, claInfos)
		if err1 != nil {
			cla.log.FuncError(cla.ClaRepo.DB.SaveClassInfosToDB, err1)
			return errcode.ErrCourseSave
		}
		err2 := cla.Sac.DB.SaveManyStudentAndCourseToDB(ctx, scs)
		if err2 != nil {
			cla.log.FuncError(cla.Sac.DB.SaveManyStudentAndCourseToDB, err2)
			return errcode.ErrCourseSave
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}

	go func() {

		//缓存
		//如果保存时其value为NULL,则直接覆盖
		err := cla.ClaRepo.Cache.OnlyAddClassInfosToCache(context.Background(),
			GenerateClassInfosKey(stuId, xnm, xqm),
			claInfos)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.OnlyAddClassInfosToCache, err)
		}
		var classIds = make([]string, 0)
		var ClassInfoKeys = make([]string, 0)
		ScKey := GenerateScSetName(stuId, xnm, xqm)
		for _, v := range claInfos {
			classIds = append(classIds, v.ID)
			ClassInfoKeys = append(ClassInfoKeys, GenerateClassInfoKey(v.ID))
		}
		//添加所有单个课程信息到缓存
		err = cla.ClaRepo.Cache.SaveManyClassInfosToCache(context.Background(), ClassInfoKeys, claInfos)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.SaveManyClassInfosToCache, err)
		}
		//保存学生ID与课程ID的对应关系到缓存
		err = cla.Sac.Cache.SaveManyStudentAndCourseToCache(context.Background(), ScKey, classIds)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.SaveManyStudentAndCourseToCache, err)
		}
	}()
	return nil
}

func (cla ClassRepo) GetAllClasses(ctx context.Context, stuId, xnm, xqm string) ([]*ClassInfo, error) {
	cacheGet := true
	key := GenerateClassInfosKey(stuId, xnm, xqm)
	classInfos, err := cla.ClaRepo.Cache.GetClassInfosFromCache(ctx, key)
	//如果err!=nil(err==redis.Nil)说明该ID第一次进入（redis中没有这个KEY），且未经过数据库，则允许其查数据库，所以要设置cacheGet=false
	//如果err==nil说明其至少经过数据库了，redis中有这个KEY,但可能值为NULL，如果不为NULL，就说明缓存命中了,直接返回没有问题
	//如果为NULL，就说明数据库中没有的数据，其依然在请求，会影响数据库（缓存穿透），我们依然直接返回
	//这时我们就需要直接返回redis中的null，即直接返回nil,而不经过数据库

	if err != nil {
		cacheGet = false
		cla.log.FuncError(cla.ClaRepo.Cache.GetClassInfosFromCache, err)
	}
	if !cacheGet {
		//从数据库中获取并用缓存加速
		var claIds []string
		claIds, err = cla.Sac.Cache.GetClassIdsFromCache(ctx, GenerateScSetName(stuId, xnm, xqm))
		if err != nil || len(claIds) == 0 {
			if err != nil {
				cla.log.FuncError(cla.Sac.Cache.GetClassIdsFromCache, err)
			}
			claIds, err = cla.Sac.DB.GetClassIDsFromSCInDB(ctx, stuId, xnm, xqm)
			if err != nil {
				if err != nil {
					cla.log.FuncError(cla.Sac.DB.GetClassIDsFromSCInDB, err)
				}
				return nil, errcode.ErrClassNotFound
			}
		}
		for _, Id := range claIds {
			classInfo, err := cla.ClaRepo.Cache.GetClassInfoFromCache(ctx, GenerateClassInfoKey(Id))
			if err != nil || classInfo == nil {
				if err != nil {
					cla.log.FuncError(cla.ClaRepo.Cache.GetClassInfoFromCache, err)
				}
				classInfo, err = cla.ClaRepo.DB.GetClassInfoFromDB(ctx, Id)
				if err != nil {
					cla.log.FuncError(cla.ClaRepo.DB.GetClassInfoFromDB, err)
					return nil, errcode.ErrClassNotFound
				}
			}
			classInfos = append(classInfos, classInfo)
		}
		go func() {
			//将课程信息当作整体存入redis
			//注意:如果未获取到，即classInfos为nil，redis仍然会设置key-value，只不过value为NULL
			err := cla.ClaRepo.Cache.OnlyAddClassInfosToCache(context.Background(), key, classInfos)
			if err != nil {
				cla.log.FuncError(cla.ClaRepo.Cache.OnlyAddClassInfosToCache, err)
			}
			//将学号与课程的ID存入缓存
			//若classIds为nil,不会进入for循环，即没有经过redis
			//下面的存取各个课程信息也同理，不会进入redis
			err = cla.Sac.Cache.SaveManyStudentAndCourseToCache(context.Background(),
				GenerateScSetName(stuId, xnm, xqm),
				claIds)
			if err != nil {
				cla.log.FuncError(cla.Sac.Cache.SaveManyStudentAndCourseToCache, err)
			}
			//同时存取各个课程信息
			for _, classInfo := range classInfos {
				key1 := GenerateClassInfoKey(classInfo.ID)
				err = cla.ClaRepo.Cache.OnlyAddClassInfoToCache(context.Background(), key1, classInfo)
				if err != nil {
					cla.log.FuncError(cla.ClaRepo.Cache.OnlyAddClassInfoToCache, err)
				}
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
	return classInfos, nil
}
func (cla ClassRepo) GetSpecificClassInfo(ctx context.Context, classId string) (*ClassInfo, error) {
	classInfo, err := cla.ClaRepo.Cache.GetClassInfoFromCache(ctx, classId)
	if err != nil {
		cla.log.FuncError(cla.ClaRepo.Cache.GetClassInfoFromCache, err)
		classInfo, err = cla.ClaRepo.DB.GetClassInfoFromDB(ctx, classId)
		if err != nil || classInfo == nil {
			cla.log.FuncError(cla.ClaRepo.DB.GetClassInfoFromDB, err)
			return nil, errcode.ErrClassNotFound
		}
		go func() {
			// 缓存
			err := cla.ClaRepo.Cache.OnlyAddClassInfoToCache(ctx, GenerateClassInfoKey(classId), classInfo)
			if err != nil {
				cla.log.FuncError(cla.ClaRepo.Cache.OnlyAddClassInfoToCache, err)
			}
		}()
	}
	return classInfo, nil
}
func (cla ClassRepo) AddClass(ctx context.Context, classInfo *ClassInfo, sc *StudentCourse, xnm, xqm string) error {
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		if err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, classInfo); err != nil {
			cla.log.FuncError(cla.ClaRepo.DB.AddClassInfoToDB, err)
			return errcode.ErrClassUpdate
		}

		// 处理 StudentCourse
		if err := cla.Sac.DB.SaveStudentAndCourseToDB(ctx, sc); err != nil {
			cla.log.FuncError(cla.Sac.DB.SaveStudentAndCourseToDB, err)
			return errcode.ErrClassUpdate
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}
	// 在事务成功提交后，异步处理缓存更新
	go func() {
		// 课程信息缓存
		stuId := sc.StuID
		key1 := GenerateClassInfoKey(classInfo.ID)
		key2 := GenerateScSetName(sc.StuID, xnm, xqm)
		err := cla.ClaRepo.Cache.AddClassInfoToCache(context.Background(), key1, GenerateClassInfosKey(stuId, xnm, xqm), classInfo)
		err = cla.Sac.Cache.AddStudentAndCourseToCache(context.Background(), key2, sc.ClaID)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.AddStudentAndCourseToCache, err)
		}
	}()
	// 不等待缓存写入完成，直接返回
	return nil
}
func (cla ClassRepo) DeleteClass(ctx context.Context, classId string, stuId string, xnm string, xqm string) error {
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		err := cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, GenerateSCID(stuId, classId, xnm, xqm))
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.DeleteStudentAndCourseInDB, err)
			return errcode.ErrClassDelete
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}
	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, classId, GenerateClassInfosKey(stuId, xnm, xqm))
	if err != nil {
		cla.log.FuncError(cla.ClaRepo.Cache.DeleteClassInfoFromCache, err)
		return err
	}
	key1 := GenerateScSetName(stuId, xnm, xqm)
	key2 := GenerateRecycleSetName(stuId, xnm, xqm)
	//删除并添加进回收站
	err = cla.Sac.Cache.DeleteAndRecycleClassId(ctx, key1, key2, classId)
	if err != nil {
		cla.log.FuncError(cla.Sac.Cache.DeleteAndRecycleClassId, err)
		return err
	}
	return nil
}
func (cla ClassRepo) GetRecycledIds(ctx context.Context, stuId, xnm, xqm string) ([]string, error) {
	recycleKey := GenerateRecycleSetName(stuId, xnm, xqm)
	classIds, err := cla.Sac.Cache.GetRecycledClassIds(ctx, recycleKey)
	if err != nil {
		cla.log.FuncError(cla.Sac.Cache.GetRecycledClassIds, err)
		return nil, err
	}
	return classIds, nil
}
func (cla ClassRepo) CheckClassIdIsInRecycledBin(ctx context.Context, stuId, xnm, xqm, classId string) bool {
	RecycledBinKey := GenerateRecycleSetName(stuId, xnm, xqm)
	return cla.Sac.Cache.CheckRecycleIdIsExist(ctx, RecycledBinKey, classId)
}
func (cla ClassRepo) RemoveClassFromRecycledBin(ctx context.Context, stuId, xnm, xqm, classId string) error {
	RecycledBinKey := GenerateRecycleSetName(stuId, xnm, xqm)
	return cla.Sac.Cache.RemoveClassFromRecycledBin(ctx, RecycledBinKey, classId)
}
func (cla ClassRepo) UpdateClass(ctx context.Context, newClassInfo *ClassInfo, newSc *StudentCourse, stuId, oldClassId, xnm, xqm string) error {
	errTx := cla.TxCtrl.InTx(ctx, func(ctx context.Context) error {
		//添加新的课程信息
		err := cla.ClaRepo.DB.AddClassInfoToDB(ctx, newClassInfo)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.DB.AddClassInfoToDB, err)
			return errcode.ErrClassUpdate
		}
		//删除原本的学生与课程的对应关系
		err = cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, GenerateSCID(stuId, oldClassId, xnm, xqm))
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.DeleteStudentAndCourseInDB, err)
			return errcode.ErrClassUpdate
		}
		//添加新的对应关系
		err = cla.Sac.DB.SaveStudentAndCourseToDB(ctx, newSc)
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.SaveStudentAndCourseToDB, err)
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
		err := cla.ClaRepo.Cache.FixClassInfoInCache(context.Background(),
			oldClassId,
			GenerateClassInfoKey(newClassInfo.ID),
			GenerateClassInfosKey(stuId, xnm, xqm),
			newClassInfo)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.FixClassInfoInCache, err)
		}
		//删除老的对应关系，添加新的对应关系
		err = cla.Sac.Cache.DeleteStudentAndCourseFromCache(context.Background(), GenerateScSetName(stuId, xnm, xqm), oldClassId)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.DeleteStudentAndCourseFromCache, err)
		}
		err = cla.Sac.Cache.AddStudentAndCourseToCache(context.Background(), GenerateScSetName(stuId, xnm, xqm), newSc.ClaID)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.AddStudentAndCourseToCache, err)
		}
	}()
	return nil
}
func (cla ClassRepo) CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool {
	key := GenerateScSetName(stuId, xnm, xqm)
	exist, err := cla.Sac.Cache.CheckExists(ctx, key, classId)
	if err == nil {
		return exist
	}
	return cla.Sac.DB.CheckExists(ctx, xnm, xqm, stuId, classId)
}
func (cla ClassRepo) GetAllSchoolClassInfos(ctx context.Context, xnm, xqm string) []*ClassInfo {
	classInfos, err := cla.ClaRepo.DB.GetAllClassInfos(ctx, xnm, xqm)
	if err != nil {
		cla.log.FuncError(cla.ClaRepo.DB.GetAllClassInfos, err)
		return nil
	}
	return classInfos
}
func GenerateScSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("StuAndCla:%s:%s:%s", stuId, xnm, xqm)
}
func GenerateRecycleSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("Recycle:%s:%s:%s", stuId, xnm, xqm)
}
func GenerateClassInfosKey(stuId, xnm, xqm string) string {
	return fmt.Sprintf("ClassInfos:%s:%s:%s", stuId, xnm, xqm)
}
func GenerateClassInfoKey(classId string) string {
	return fmt.Sprintf("ClassInfo:%s", classId)
}
