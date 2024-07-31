package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"reflect"
	"runtime"
)

var ProviderSet = wire.NewSet(NewLogger)

// LogerPrinter  打印函数错误
type LogerPrinter interface {
	FuncError(f interface{}, err error)
}

type LoggerPrint struct {
	log *log.Helper
}

func NewLogger(logger log.Logger) LogerPrinter {
	return &LoggerPrint{
		log: log.NewHelper(logger),
	}
}

func (l *LoggerPrint) FuncError(f interface{}, err error) {
	// 使用reflect包获取函数的值
	value := reflect.ValueOf(f)
	name := runtime.FuncForPC(value.Pointer()).Name()
	l.log.Errorf("Err %v:%v \n", name, err)
}
