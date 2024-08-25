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
		var classIds = make([]string, 0)
		key := GenerateSetName(stuId, xnm, xqm)
		for _, v := range claInfos {
			classIds = append(classIds, v.ID)
		}
		err := cla.ClaRepo.Cache.SaveManyClassInfosToCache(ctx, classIds, claInfos)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.SaveManyClassInfosToCache, err)
		}
		err = cla.Sac.Cache.SaveManyStudentAndCourseToCache(ctx, key, classIds)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.SaveManyStudentAndCourseToCache, err)
		}
	}()
	return nil
}

func (cla ClassRepo) GetAllClasses(ctx context.Context, stuId, xnm, xqm string) ([]*ClassInfo, error) {
	var classInfos = make([]*ClassInfo, 0)
	cacheGet := true
	key1 := GenerateSetName(stuId, xnm, xqm)
	claIds, err := cla.Sac.Cache.GetClassIdsFromCache(ctx, key1)
	if err != nil || len(claIds) == 0 {
		cla.log.FuncError(cla.Sac.Cache.GetClassIdsFromCache, err)
		claIds, err = cla.Sac.DB.GetClassIDsFromSCInDB(ctx, stuId, xnm, xqm)
		if len(claIds) == 0 {
			cla.log.FuncError(cla.Sac.DB.GetClassIDsFromSCInDB, err)
			return nil, errcode.ErrClassNotFound
		}
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.GetClassIDsFromSCInDB, err)
			return nil, errcode.ErrClassFound
		}
		go func() {
			//缓存获取失败的话就再次去缓存
			if claIds == nil {
				return
			}
			err := cla.Sac.Cache.SaveManyStudentAndCourseToCache(ctx, key1, claIds)
			if err != nil {
				cla.log.FuncError(cla.Sac.Cache.SaveManyStudentAndCourseToCache, err)
			}
		}()
	}
	for _, classId := range claIds {
		key := classId
		classInfo, err := cla.ClaRepo.Cache.GetClassInfoFromCache(ctx, key)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.GetClassInfoFromCache, err)
			cacheGet = false
			classInfos = classInfos[:0]
			break
		}
		classInfos = append(classInfos, classInfo)
	}
	if !cacheGet {
		for _, Id := range claIds {
			classInfo, err := cla.ClaRepo.DB.GetClassInfoFromDB(ctx, Id)
			if err != nil {
				cla.log.FuncError(cla.ClaRepo.DB.GetClassInfoFromDB, err)
				return nil, errcode.ErrClassNotFound
			}
			classInfos = append(classInfos, classInfo)
		}
		go func() {
			//缓存
			var classIds = make([]string, 0)
			for _, v := range classInfos {
				classIds = append(classIds, v.ID)
			}
			err := cla.ClaRepo.Cache.SaveManyClassInfosToCache(ctx, classIds, classInfos)
			if err != nil {
				cla.log.FuncError(cla.ClaRepo.Cache.SaveManyClassInfosToCache, err)
			}
		}()
	}
	return classInfos, nil
}
func (cla ClassRepo) GetSpecificClassInfo(ctx context.Context, classId string) (*ClassInfo, error) {
	classInfo, err := cla.ClaRepo.Cache.GetClassInfoFromCache(ctx, classId)
	if err != nil {
		cla.log.FuncError(cla.ClaRepo.Cache.GetClassInfoFromCache, err)
		classInfo, err = cla.ClaRepo.DB.GetClassInfoFromDB(ctx, classId)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.DB.GetClassInfoFromDB, err)
			return nil, errcode.ErrClassNotFound
		}
		go func() {
			// 缓存
			err := cla.ClaRepo.Cache.AddClassInfoToCache(ctx, classId, classInfo)
			if err != nil {
				cla.log.FuncError(cla.ClaRepo.Cache.AddClassInfoToCache, err)
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
		key1 := classInfo.ID
		key2 := GenerateSetName(sc.StuID, xnm, xqm)
		err := cla.ClaRepo.Cache.AddClassInfoToCache(ctx, key1, classInfo)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.AddClassInfoToCache, err)
		}
		err = cla.Sac.Cache.AddStudentAndCourseToCache(ctx, key2, sc.ClaID)
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

	////删除缓存
	//err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, classId)
	//if err != nil {
	//	cla.log.FuncError(cla.ClaRepo.Cache.DeleteClassInfoFromCache, err)
	//	return errcode.ErrClassDelete
	//}
	key := GenerateSetName(stuId, xnm, xqm)
	err := cla.Sac.Cache.DeleteStudentAndCourseFromCache(ctx, key, classId)
	if err != nil {
		cla.log.FuncError(cla.Sac.Cache.DeleteStudentAndCourseFromCache, err)
		return errcode.ErrClassDelete
	}
	return nil
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
		err := cla.ClaRepo.Cache.AddClassInfoToCache(ctx, newClassInfo.ID, newClassInfo)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.Cache.AddClassInfoToCache, err)
		}
		err = cla.Sac.Cache.DeleteStudentAndCourseFromCache(ctx, GenerateSetName(stuId, xnm, xqm), oldClassId)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.DeleteStudentAndCourseFromCache, err)
		}
		err = cla.Sac.Cache.AddStudentAndCourseToCache(ctx, GenerateSetName(stuId, xnm, xqm), newSc.ClaID)
		if err != nil {
			cla.log.FuncError(cla.Sac.Cache.AddStudentAndCourseToCache, err)
		}
	}()
	return nil
}
func (cla ClassRepo) CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool {
	key := GenerateSetName(stuId, xnm, xqm)
	exist, err := cla.Sac.Cache.CheckExists(ctx, key, classId)
	if err == nil {
		return exist
	}
	return cla.Sac.DB.CheckExists(ctx, xnm, xqm, stuId, classId)
}
func (cla ClassRepo) GetAllSchoolClassInfos(ctx context.Context) []*ClassInfo {
	clasInfos := make([]*ClassInfo, 0)
	classids, err := cla.Sac.DB.GetAllSchoolClassIds(ctx)
	if err != nil {
		cla.log.FuncError(cla.Sac.DB.GetAllSchoolClassIds, err)
		return nil
	}
	newclassIds := removeDuplicates(classids)
	for _, classId := range newclassIds {
		clasInfo, err := cla.ClaRepo.DB.GetClassInfoFromDB(ctx, classId)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.DB.GetClassInfoFromDB, err)
			continue
		}
		clasInfos = append(clasInfos, clasInfo)
	}
	return clasInfos
}
func GenerateSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("StuAndCla:%s:%s:%s", stuId, xnm, xqm)
}

// 去重
func removeDuplicates(strSlice []string) []string {
	// 创建一个空的 map 来跟踪已经存在的字符串
	uniqueMap := make(map[string]bool)
	// 创建一个空的切片来存储去重后的结果
	result := []string{}

	// 遍历输入的字符串切片
	for _, str := range strSlice {
		// 如果字符串不在 map 中，说明是唯一的
		if _, exists := uniqueMap[str]; !exists {
			// 将字符串加入结果切片
			result = append(result, str)
			// 并在 map 中标记该字符串已经存在
			uniqueMap[str] = true
		}
	}

	return result
}
