package biz

import (
	"class/internal/errcode"
	log "class/internal/logPrinter"
	"context"
	"fmt"
	"regexp"
	"strconv"
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
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.GetClassIDsFromSCInDB, err)
			return nil, errcode.ErrClassNotFound
		}
		go func() {
			//缓存获取失败的话就再次去缓存
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
		err := cla.ClaRepo.DB.DeleteClassInfoInDB(ctx, classId)
		if err != nil {
			cla.log.FuncError(cla.ClaRepo.DB.DeleteClassInfoInDB, err)
			return errcode.ErrClassDelete
		}
		err = cla.Sac.DB.DeleteStudentAndCourseInDB(ctx, GenerateSCID(stuId, classId, xnm, xqm))
		if err != nil {
			cla.log.FuncError(cla.Sac.DB.DeleteStudentAndCourseInDB, err)
			return errcode.ErrClassDelete
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}

	//删除缓存
	err := cla.ClaRepo.Cache.DeleteClassInfoFromCache(ctx, classId)
	if err != nil {
		cla.log.FuncError(cla.ClaRepo.Cache.DeleteClassInfoFromCache, err)
		return errcode.ErrClassDelete
	}
	key := GenerateSetName(stuId, xnm, xqm)
	err = cla.Sac.Cache.DeleteStudentAndCourseFromCache(ctx, key, classId)
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
func GenerateSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("StuAndCla:%s:%s:%s", stuId, xnm, xqm)
}
func Check(id string, day int64, dur string) bool {
	day1, dur1, err := ExtractDayAndClassWhen(id)
	if err != nil {
		return false
	}
	if day != day1 || dur != dur1 {
		return false
	}
	return true
}

// ExtractDayAndClassWhen 提取格式化字符串中的 day 和 classwhen
func ExtractDayAndClassWhen(id string) (int64, string, error) {
	// 定义正则表达式来匹配 day 和 classwhen
	re := regexp.MustCompile(`^Class:\w+:\w+:\w+:(\d+):(\w+):`)

	// 找到匹配的子字符串
	matches := re.FindStringSubmatch(id)
	if len(matches) < 3 {
		return 0, "", fmt.Errorf("could not extract day and classwhen from ID: %s", id)
	}

	// 将 day 转换为 int
	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("error converting day to int: %v", err)
	}

	classwhen := matches[2]
	return int64(day), classwhen, nil
}
