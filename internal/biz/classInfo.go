package biz

import "context"

type ClassInfoDBRepo interface {
	SaveClassInfosToDB(ctx context.Context, classInfo []*ClassInfo) error
	AddClassInfoToDB(ctx context.Context, classInfo *ClassInfo) error
	GetClassInfoFromDB(ctx context.Context, ID string) (*ClassInfo, error)
	DeleteClassInfoInDB(ctx context.Context, ID string) error
	//UpdateClassInfoInDB(ctx context.Context, tx *gorm.DB, classInfo *ClassInfo) error
}
type ClassInfoCacheRepo interface {
	SaveManyClassInfosToCache(ctx context.Context, keys []string, classInfos []*ClassInfo) error
	AddClassInfoToCache(ctx context.Context, key string, classInfo *ClassInfo) error
	GetClassInfoFromCache(ctx context.Context, key string) (*ClassInfo, error)
	DeleteClassInfoFromCache(ctx context.Context, key string) error
	UpdateClassInfoInCache(ctx context.Context, key string, classInfo *ClassInfo) error
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