package errcode

import (
	v1 "class/api/classer/v1"
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrClassNotFound = errors.New(200, v1.ErrorReason_DB_NOTFOUND.String(), "课程信息未找到")
	ErrClassFound    = errors.New(450, v1.ErrorReason_DB_FINDERR.String(), "数据库查找课程失败")
	ErrClassUpdate   = errors.New(300, v1.ErrorReason_DB_UPDATEERR.String(), "课程更新失败")
	ErrParam         = errors.New(301, v1.ErrorReason_DB_UPDATEERR.String(), "入参错误")
	ErrCourseSave    = errors.New(302, v1.ErrorReason_DB_SAVEERROR.String(), "课程保存失败")
	ErrClassDelete   = errors.New(303, v1.ErrorReason_DB_DELETEERROR.String(), "课程删除失败")
	ErrCrawler       = errors.New(304, v1.ErrorReason_Crawler_Error.String(), "爬取课表失败")
)
