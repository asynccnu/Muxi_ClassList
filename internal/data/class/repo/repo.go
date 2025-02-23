package repo

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
)

type ClassCache interface {
	//通过缓存，来查询某个学生的所有课程ID
	GetClassIDList(ctx context.Context, stuID, year, semester string) ([]string, error)
	//通过课程ID，查询课程
	GetClassesByID(ctx context.Context, classids ...string) ([]model.ClassDO, error)
	//设置学生和课程ID的对应关系
	SetClassIDList(ctx context.Context, stuID, year, semester string, classids ...string) error
	//添加课程
	AddClass(ctx context.Context, classes ...model.ClassDO) error
	//删除缓存的对应关系
	DeleteClassIDList(ctx context.Context, stuID, year, semester string) error
	//删除课程
	DeleteClass(ctx context.Context, classID string) error
}
