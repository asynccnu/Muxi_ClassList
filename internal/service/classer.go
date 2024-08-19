package service

import (
	pb "class/api/classer/v1"
	"class/internal/biz"
	"class/internal/errcode"
	"class/internal/logPrinter"
	"class/internal/pkg/tool"
	"context"
)

//go:generate mockgen -source=./classer.go -destination=./mock/mock_classer.go -package=mock_service
type ClassCtrl interface {
	CheckSCIdsExist(ctx context.Context, stuId, classId, xnm, xqm string) bool
	GetClasses(ctx context.Context, StuId string, week int64, xnm, xqm string, cookie string) ([]*biz.Class, error)
	AddClass(ctx context.Context, stuId string, info *biz.ClassInfo) error
	DeleteClass(ctx context.Context, classId string, stuId string, xnm string, xqm string) error
	SearchClass(ctx context.Context, classId string) (*biz.ClassInfo, error)
	UpdateClass(ctx context.Context, newClassInfo *biz.ClassInfo, newSc *biz.StudentCourse, stuId, oldClassId, xnm, xqm string) error
}
type CCNUServiceProxy interface {
	GetCookie(ctx context.Context, stu string) (string, error)
}
type ClasserService struct {
	pb.UnimplementedClasserServer
	Clu ClassCtrl
	Cs  CCNUServiceProxy
	log logPrinter.LogerPrinter
}

func NewClasserService(clu ClassCtrl, cs CCNUServiceProxy, log logPrinter.LogerPrinter) *ClasserService {
	return &ClasserService{
		Clu: clu,
		Cs:  cs,
		log: log,
	}
}

func (s *ClasserService) GetClass(ctx context.Context, req *pb.GetClassRequest) (*pb.GetClassResponse, error) {
	cookie, err := s.Cs.GetCookie(ctx, req.GetStuId())
	if err != nil {
		s.log.FuncError(s.Cs.GetCookie, err)
	}
	//调试专用
	//cookie := "JSESSIONID=E48CAEEB7D2EA3CF0ABE01546CCCDE13"
	pclasses := make([]*pb.Class, 0)

	if !tool.CheckSY(req.Semester, req.Year) || req.GetWeek() <= 0 {
		return &pb.GetClassResponse{}, errcode.ErrParam
	}

	classes, err := s.Clu.GetClasses(ctx, req.GetStuId(), req.GetWeek(), req.GetYear(), req.GetSemester(), cookie)
	if err != nil {
		s.log.FuncError(s.Clu.GetClasses, err)
		return &pb.GetClassResponse{}, err
	}
	for _, class := range classes {
		pinfo := HandleClass(class.Info)
		var pclass = &pb.Class{
			Info:     pinfo,
			Thisweek: class.ThisWeek,
		}
		pclasses = append(pclasses, pclass)
	}
	return &pb.GetClassResponse{
		Classes: pclasses,
	}, nil
}
func (s *ClasserService) AddClass(ctx context.Context, req *pb.AddClassRequest) (*pb.AddClassResponse, error) {
	if !tool.CheckSY(req.Semester, req.Year) || req.GetWeeks() <= 0 {
		return &pb.AddClassResponse{}, errcode.ErrParam
	}
	weekDur := tool.FormatWeeks(tool.ParseWeeks(req.Weeks))
	var classInfo = &biz.ClassInfo{
		Day:          req.GetDay(),
		Teacher:      req.GetTeacher(),
		Where:        req.GetWhere(),
		ClassWhen:    req.GetDurClass(),
		WeekDuration: weekDur,
		Classname:    req.GetName(),
		Credit:       req.GetCredit(),
		Weeks:        req.GetWeeks(),
		Semester:     req.GetSemester(),
		Year:         req.GetYear(),
	}
	classInfo.UpdateID()
	err := s.Clu.AddClass(ctx, req.GetStuId(), classInfo)
	if err != nil {
		s.log.FuncError(s.Clu.AddClass, err)
		return &pb.AddClassResponse{}, err
	}

	return &pb.AddClassResponse{
		Id:  classInfo.ID,
		Msg: "成功添加",
	}, nil
}
func (s *ClasserService) DeleteClass(ctx context.Context, req *pb.DeleteClassRequest) (*pb.DeleteClassResponse, error) {
	exist := s.Clu.CheckSCIdsExist(ctx, req.GetStuId(), req.GetId(), req.GetYear(), req.GetSemester())
	if !exist {
		return &pb.DeleteClassResponse{
			Msg: "该课程不存在",
		}, errcode.ErrSCIDNOTEXIST
	}
	err := s.Clu.DeleteClass(ctx, req.GetId(), req.GetStuId(), req.GetYear(), req.GetSemester())
	if err != nil {
		s.log.FuncError(s.Clu.DeleteClass, err)
		return &pb.DeleteClassResponse{}, err
	}
	return &pb.DeleteClassResponse{
		Msg: "成功删除",
	}, nil
}
func (s *ClasserService) UpdateClass(ctx context.Context, req *pb.UpdateClassRequest) (*pb.UpdateClassResponse, error) {
	exist := s.Clu.CheckSCIdsExist(ctx, req.GetStuId(), req.GetClassId(), req.GetYear(), req.GetSemester())
	if !exist {
		return &pb.UpdateClassResponse{
			Msg: "该课程不存在",
		}, errcode.ErrSCIDNOTEXIST
	}
	if !tool.CheckSY(req.Semester, req.GetYear()) || req.GetWeeks() <= 0 {
		return &pb.UpdateClassResponse{}, errcode.ErrParam
	}
	weekDur := tool.FormatWeeks(tool.ParseWeeks(req.GetWeeks()))
	oldclassInfo, err := s.Clu.SearchClass(ctx, req.GetClassId())
	if err != nil {
		s.log.FuncError(s.Clu.SearchClass, err)
		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	oldclassInfo.Day = req.GetDay()
	oldclassInfo.Teacher = req.GetTeacher()
	oldclassInfo.Where = req.GetWhere()
	oldclassInfo.ClassWhen = req.GetDurClass()
	oldclassInfo.WeekDuration = weekDur
	oldclassInfo.Classname = req.GetName()
	oldclassInfo.Weeks = req.GetWeeks()
	oldclassInfo.UpdateID()
	newSc := &biz.StudentCourse{
		StuID:           req.GetStuId(),
		ClaID:           oldclassInfo.ID,
		Year:            oldclassInfo.Year,
		Semester:        oldclassInfo.Semester,
		IsManuallyAdded: false,
	}
	newSc.UpdateID()
	err = s.Clu.UpdateClass(ctx, oldclassInfo, newSc, req.GetStuId(), req.GetClassId(), req.GetYear(), req.GetSemester())
	if err != nil {
		s.log.FuncError(s.Clu.UpdateClass, err)
		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	return &pb.UpdateClassResponse{
		ClassId: oldclassInfo.ID,
		Msg:     "成功修改",
	}, nil
}
func HandleClass(info *biz.ClassInfo) *pb.ClassInfo {
	return &pb.ClassInfo{
		Day:          info.Day,
		Teacher:      info.Teacher,
		Where:        info.Where,
		ClassWhen:    info.ClassWhen,
		WeekDuration: info.WeekDuration,
		Classname:    info.Classname,
		Credit:       info.Credit,
		Weeks:        info.Weeks,
		Id:           info.ID,
		Semester:     info.Semester,
		Year:         info.Year,
	}
}
