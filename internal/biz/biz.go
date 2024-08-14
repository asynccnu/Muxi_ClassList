package biz

import (
	"context"
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
