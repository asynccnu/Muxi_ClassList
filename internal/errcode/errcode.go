package errcode

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewErr(code int, msg string) Err {
	return Err{Code: code, Msg: msg}
}

func (e Err) Error() string {
	return e.Msg
}

var (
	ErrClassNotFound         = NewErr(450, "课程信息未找到")
	ErrClassFound            = NewErr(451, "数据库查找课程失败")
	ErrClassUpdate           = NewErr(452, "课程更新失败")
	ErrParam                 = NewErr(453, "入参错误")
	ErrCourseSave            = NewErr(454, "课程保存失败")
	ErrClassDelete           = NewErr(455, "课程删除失败")
	ErrCrawler               = NewErr(456, "爬取课表失败")
	ErrCCNULogin             = NewErr(457, "请求ccnu一站式登录服务错误")
	ErrSCIDNOTEXIST          = NewErr(458, "学号与课程ID的对应关系未找到")
	ErrRecycleBinDoNotHaveIt = NewErr(459, "回收站中不存在该课程")
	ErrRecover               = NewErr(460, "恢复课程失败")
	ErrGetStuIdByJxbId       = NewErr(461, "通过jxb_id获取stu_ids获取失败")
	ErrClassIsExist          = NewErr(462, "已有该课程")
	ErrClassAdd              = NewErr(463, "添加课程失败")
	ErrGetRecycledClasses    = NewErr(464, "获取回收站课程失败")
)
