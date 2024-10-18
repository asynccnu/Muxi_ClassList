package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
)

type ClassInfoDBRepo interface {
	SaveClassInfosToDB(ctx context.Context, classInfo []*model.ClassInfo) error
	AddClassInfoToDB(ctx context.Context, classInfo *model.ClassInfo) error
	GetClassInfoFromDB(ctx context.Context, ID string) (*model.ClassInfo, error)
	DeleteClassInfoInDB(ctx context.Context, ID string) error
	GetClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*model.ClassInfo, error)
}

type ClassInfoCacheRepo interface {
	OnlyAddClassInfosToCache(ctx context.Context, key string, classInfos []*model.ClassInfo) error
	GetClassInfoFromCache(ctx context.Context, key string) (*model.ClassInfo, error)
	GetClassInfosFromCache(ctx context.Context, key string) ([]*model.ClassInfo, error)
	DeleteClassInfoFromCache(ctx context.Context, deletedId, classInfosKey string) error
	UpdateClassInfoInCache(ctx context.Context, oldID, classInfosKey string, classInfo *model.ClassInfo, add bool) error
}
type ClassInfoRepo struct {
	DB    ClassInfoDBRepo
	Cache ClassInfoCacheRepo
}

func NewClassInfoRepo(DB ClassInfoDBRepo, Cache ClassInfoCacheRepo) *ClassInfoRepo {
	return &ClassInfoRepo{
		DB:    DB,
		Cache: Cache,
	}
}
