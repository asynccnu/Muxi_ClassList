package biz

import (
	"context"
	model2 "github.com/asynccnu/Muxi_ClassList/internal/model"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewClassUsercase, NewClassInfoRepo, NewStudentAndCourseRepo, NewClassRepo)

type Transaction interface {
	// 下面2个方法配合使用，在InTx方法中执行ORM操作的时候需要使用DB方法获取db！
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
	DB(ctx context.Context) *gorm.DB
}
type ClassCrawler interface {
	GetClassInfosForUndergraduate(ctx context.Context, req model2.GetClassInfosForUndergraduateReq) (*model2.GetClassInfosForUndergraduateResp, error)
	GetClassInfoForGraduateStudent(ctx context.Context, req model2.GetClassInfoForGraduateStudentReq) (*model2.GetClassInfoForGraduateStudentResp, error)
}
type ClassRepoProxy interface {
	//保存课程
	SaveClass(ctx context.Context, stuID, year, semester string, classInfos []*model2.ClassInfo, scs []*model2.StudentCourse)
	//获取某个学生某个学期的所有课程
	GetAllClasses(ctx context.Context, req model2.GetAllClassesReq) (*model2.GetAllClassesResp, error)
	//只获取特定ID的class_info
	GetSpecificClassInfo(ctx context.Context, req model2.GetSpecificClassInfoReq) (*model2.GetSpecificClassInfoResp, error)
	//添加课程
	AddClass(ctx context.Context, req model2.AddClassReq) error
	//删除课程
	DeleteClass(ctx context.Context, req model2.DeleteClassReq) error
	//获取某个学生某个学期的处于回收站的课程ID
	GetRecycledIds(ctx context.Context, req model2.GetRecycledIdsReq) (*model2.GetRecycledIdsResp, error)
	//恢复课程
	RecoverClassFromRecycledBin(ctx context.Context, req model2.RecoverClassFromRecycleBinReq) error
	//更新课程
	UpdateClass(ctx context.Context, req model2.UpdateClassReq) error
	//判断课程和学生ID是否有联系
	CheckSCIdsExist(ctx context.Context, req model2.CheckSCIdsExistReq) bool
	//获取全校某个学期的所有课程
	GetAllSchoolClassInfos(ctx context.Context, req model2.GetAllSchoolClassInfosReq) *model2.GetAllSchoolClassInfosResp
	//检查某个class是否存在于回收站中
	CheckClassIdIsInRecycledBin(ctx context.Context, req model2.CheckClassIdIsInRecycledBinReq) bool
	//获取某个学生某个学期的手动添加的课程[直接来自数据库]
	GetAddedClasses(ctx context.Context, req model2.GetAddedClassesReq) (*model2.GetAddedClassesResp, error)
}
type JxbRepo interface {
	SaveJxb(ctx context.Context, stuID string, jxbID []string) error
	FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
type CCNUServiceProxy interface {
	GetCookie(ctx context.Context, stuID string) (string, error)
}
