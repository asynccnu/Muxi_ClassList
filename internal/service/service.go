package service

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewClasserService)

type ClassCtrl interface {
	GetClasses(ctx context.Context, stuID, year, semester string, week int64) ([]*model.Class, error)
	AddClass(ctx context.Context, stuID string, info *model.ClassInfo) error
	DeleteClass(ctx context.Context, stuID, year, semester, classId string) error
	GetRecycledClassInfos(ctx context.Context, stuID, year, semester string) ([]*model.ClassInfo, error)
	RecoverClassInfo(ctx context.Context, stuID, year, semester, classId string) error
	SearchClass(ctx context.Context, classId string) (*model.ClassInfo, error)
	UpdateClass(ctx context.Context, stuID, year, semester string, newClassInfo *model.ClassInfo, newSc *model.StudentCourse, oldClassId string) error
	CheckSCIdsExist(ctx context.Context, stuID, year, semester, classId string) bool
	GetAllSchoolClassInfosToOtherService(ctx context.Context, year, semester string) []*model.ClassInfo
	GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
