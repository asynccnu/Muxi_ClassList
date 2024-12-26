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
	CheckAndSaveClass(ctx context.Context, stuID, year, semester string, classInfos []*model2.ClassInfo, scs []*model2.StudentCourse)
	SaveClasses(ctx context.Context, req model2.SaveClassReq) error
	GetAllClasses(ctx context.Context, req model2.GetAllClassesReq) (*model2.GetAllClassesResp, error)
	GetSpecificClassInfo(ctx context.Context, req model2.GetSpecificClassInfoReq) (*model2.GetSpecificClassInfoResp, error)
	AddClass(ctx context.Context, req model2.AddClassReq) error
	DeleteClass(ctx context.Context, req model2.DeleteClassReq) error
	GetRecycledIds(ctx context.Context, req model2.GetRecycledIdsReq) (*model2.GetRecycledIdsResp, error)
	RecoverClassFromRecycledBin(ctx context.Context, req model2.RecoverClassFromRecycleBinReq) error
	UpdateClass(ctx context.Context, req model2.UpdateClassReq) error
	CheckSCIdsExist(ctx context.Context, req model2.CheckSCIdsExistReq) bool
	GetAllSchoolClassInfos(ctx context.Context, req model2.GetAllSchoolClassInfosReq) *model2.GetAllSchoolClassInfosResp
	CheckClassIdIsInRecycledBin(ctx context.Context, req model2.CheckClassIdIsInRecycledBinReq) bool
}
type JxbRepo interface {
	SaveJxb(ctx context.Context, stuID string, jxbID []string) error
	FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
type CCNUServiceProxy interface {
	GetCookie(ctx context.Context, stuID string) (string, error)
}
