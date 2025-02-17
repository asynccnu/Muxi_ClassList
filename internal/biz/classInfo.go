package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
)

type ClassInfoDBRepo interface {
	SaveClassInfosToDB(ctx context.Context, classInfo []*model.ClassInfo) error
	AddClassInfoToDB(ctx context.Context, classInfo *model.ClassInfo) error
	GetClassInfoFromDB(ctx context.Context, ID string) (*model.ClassInfo, error)
	GetClassInfos(ctx context.Context, stuId, xnm, xqm string) ([]*model.ClassInfo, error)
}

type ClassInfoCacheRepo interface {
	AddClaInfosToCache(ctx context.Context, key string, classInfos []*model.ClassInfo) error
	GetClassInfosFromCache(ctx context.Context, key string) ([]*model.ClassInfo, error)
	DeleteClassInfoFromCache(ctx context.Context, classInfosKey ...string) error
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
