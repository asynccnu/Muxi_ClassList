package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/google/wire"
	"time"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewClassUsecase)

type ClassCrawler interface {
	//获取本科生的课表
	GetClassInfosForUndergraduate(ctx context.Context, year, semester, cookie string) ([]*model.ClassBiz, []string, error)
	//获取研究生的课表(未实现)
	GetClassInfoForGraduateStudent(ctx context.Context, year, semester, cookie string) ([]*model.ClassBiz, []string, error)
}

type JxbRepo interface {
	//保存教学班
	SaveJxb(ctx context.Context, stuID string, jxbID []string) error
	//根据教学班ID查询stuID
	FindStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}
type CCNUServiceProxy interface {
	//从其他服务获取cookie
	GetCookie(ctx context.Context, stuID string) (string, error)
}

// 核心课程存储（必需能力）
type ClassStorage interface {
	//保存课程
	SaveClass(ctx context.Context, stuID, year, semester string, classes []*model.ClassBiz) error
	//获取某个学生某个学期的所有课程
	GetClassesFromLocal(ctx context.Context, stuID, year, semester string) ([]*model.ClassBiz, error)
	//只获取特定ID的class_info
	GetSpecificClassInfo(ctx context.Context, classID string) (*model.ClassBiz, error)
	//更新课程
	UpdateClass(ctx context.Context, stuID, year, semester string, oldClassID string, newClass *model.ClassBiz) error
	//删除课程
	DeleteClass(ctx context.Context, stuID, year, semester string, classID string) error
}

// 回收站管理
type ClassRecycleBinManager interface {
	//获取某个学生某个学期的处于回收站的课程ID
	GetRecycledClasses(ctx context.Context, stuID, year, semester string) ([]*model.ClassBiz, error)
	//恢复课程
	RecoverClassFromRecycledBin(ctx context.Context, stuID, year, semester string, classID string) error
	//检查某个class是否存在于回收站中
	CheckClassIdIsInRecycledBin(ctx context.Context, stuID, year, semester string, classID string) bool
}

// 手动课程管理（扩展能力）
type ManualClassManager interface {
	//添加课程
	AddClass(ctx context.Context, stuID, year, semester string, class *model.ClassBiz) error
	//获取某个学生某个学期的手动添加的课程[直接来自数据库]
	GetAddedClasses(ctx context.Context, stuID, year, semester string) ([]*model.ClassBiz, error)
}

// 全校课程管理（特殊场景）
type SchoolClassExplorer interface {
	//获取全校某个学期的所有课程
	GetAllSchoolClassInfos(ctx context.Context, year, semester string, cursor time.Time) []*model.ClassBiz
}

// 关联校验（辅助能力）
type ClassAssociationValidator interface {
	//判断课程和学生ID是否有联系
	CheckSCIdsExist(ctx context.Context, stuID, year, semester string, classID string) bool
}
