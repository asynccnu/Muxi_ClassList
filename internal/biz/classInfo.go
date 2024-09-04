package biz

import "context"

//go:generate mockgen -source=./classInfo.go -destination=./mock/mock_classInfo.go -package=mock_biz

type ClassInfoDBRepo interface {
	SaveClassInfosToDB(ctx context.Context, classInfo []*ClassInfo) error
	AddClassInfoToDB(ctx context.Context, classInfo *ClassInfo) error
	GetClassInfoFromDB(ctx context.Context, ID string) (*ClassInfo, error)
	DeleteClassInfoInDB(ctx context.Context, ID string) error
	GetAllClassInfos(ctx context.Context, xnm, xqm string) ([]*ClassInfo, error)
	//UpdateClassInfoInDB(ctx context.Context, tx *gorm.DB, classInfo *ClassInfo) error
}

type ClassInfoCacheRepo interface {
	SaveManyClassInfosToCache(ctx context.Context, keys []string, classInfos []*ClassInfo) error
	OnlyAddClassInfoToCache(ctx context.Context, key string, classInfo *ClassInfo) error
	OnlyAddClassInfosToCache(ctx context.Context, key string, classInfos []*ClassInfo) error
	AddClassInfoToCache(ctx context.Context, classInfoKey, classInfosKey string, classInfo *ClassInfo) error
	GetClassInfoFromCache(ctx context.Context, key string) (*ClassInfo, error)
	GetClassInfosFromCache(ctx context.Context, key string) ([]*ClassInfo, error)
	DeleteClassInfoFromCache(ctx context.Context, deletedId, classInfosKey string) error
	FixClassInfoInCache(ctx context.Context, oldID, classInfoKey, classInfosKey string, classInfo *ClassInfo) error
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
