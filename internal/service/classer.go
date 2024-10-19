package service

import (
	"context"
	pb "github.com/asynccnu/Muxi_ClassList/api/classer/v1"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

type ClassCtrl interface {
	CheckSCIdsExist(ctx context.Context, classid string) bool
	GetClasses(ctx context.Context, week int64) ([]*model.Class, error)
	AddClass(ctx context.Context, info *model.ClassInfo) error
	DeleteClass(ctx context.Context, classId string) error
	SearchClass(ctx context.Context, classId string) (*model.ClassInfo, error)
	UpdateClass(ctx context.Context, newClassInfo *model.ClassInfo, newSc *model.StudentCourse, oldClassId string) error
	GetAllSchoolClassInfosToOtherService(ctx context.Context) []*model.ClassInfo
	GetRecycledClassInfos(ctx context.Context) ([]*model.ClassInfo, error)
	RecoverClassInfo(ctx context.Context, classId string) error
	GetStuIdsByJxbId(ctx context.Context, jxbId string) ([]string, error)
}

type ClasserService struct {
	pb.UnimplementedClasserServer
	Clu ClassCtrl
	log *log.Helper
}

func NewClasserService(clu ClassCtrl, logger log.Logger) *ClasserService {
	return &ClasserService{
		Clu: clu,
		log: log.NewHelper(logger),
	}
}

func (s *ClasserService) GetClass(ctx context.Context, req *pb.GetClassRequest) (*pb.GetClassResponse, error) {
	if !tool.CheckSY(req.Semester, req.Year) || req.GetWeek() <= 0 {
		return &pb.GetClassResponse{}, errcode.ErrParam
	}
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	pclasses := make([]*pb.Class, 0)
	classes, err := s.Clu.GetClasses(ctx, req.GetWeek())
	if err != nil {
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
	//fmt.Println("getclass past:", time.Now().Sub(time2))
	return &pb.GetClassResponse{
		Classes: pclasses,
	}, nil
}
func (s *ClasserService) AddClass(ctx context.Context, req *pb.AddClassRequest) (*pb.AddClassResponse, error) {
	if !tool.CheckSY(req.Semester, req.Year) || req.GetWeeks() <= 0 || !tool.CheckIfThisYear(req.Year, req.Semester) {
		return &pb.AddClassResponse{}, errcode.ErrParam
	}
	weekDur := tool.FormatWeeks(tool.ParseWeeks(req.Weeks))
	var classInfo = &model.ClassInfo{
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
		JxbId:        "unavailable",
	}
	if req.Credit != nil {
		classInfo.Credit = req.GetCredit()
	}
	classInfo.UpdateID()
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	err := s.Clu.AddClass(ctx, classInfo)
	if err != nil {

		return &pb.AddClassResponse{}, err
	}

	return &pb.AddClassResponse{
		Id:  classInfo.ID,
		Msg: "成功添加",
	}, nil
}
func (s *ClasserService) DeleteClass(ctx context.Context, req *pb.DeleteClassRequest) (*pb.DeleteClassResponse, error) {
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	exist := s.Clu.CheckSCIdsExist(ctx, req.GetId())
	if !exist {
		return &pb.DeleteClassResponse{
			Msg: "该课程不存在",
		}, errcode.ErrSCIDNOTEXIST
	}
	err := s.Clu.DeleteClass(ctx, req.GetId())
	if err != nil {

		return &pb.DeleteClassResponse{}, err
	}
	return &pb.DeleteClassResponse{
		Msg: "成功删除",
	}, nil
}
func (s *ClasserService) UpdateClass(ctx context.Context, req *pb.UpdateClassRequest) (*pb.UpdateClassResponse, error) {
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	exist := s.Clu.CheckSCIdsExist(ctx, req.GetClassId())
	if !exist {
		return &pb.UpdateClassResponse{
			Msg: "该课程不存在",
		}, errcode.ErrSCIDNOTEXIST
	}
	if !tool.CheckSY(req.Semester, req.GetYear()) {
		return &pb.UpdateClassResponse{}, errcode.ErrParam
	}

	oldclassInfo, err := s.Clu.SearchClass(ctx, req.GetClassId())
	if err != nil {

		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	if req.Day != nil {
		oldclassInfo.Day = req.GetDay()
	}
	if req.Teacher != nil {
		oldclassInfo.Teacher = req.GetTeacher()
	}
	if req.Where != nil {
		oldclassInfo.Where = req.GetWhere()
	}
	if req.DurClass != nil {
		oldclassInfo.ClassWhen = req.GetDurClass()
	}
	//oldclassInfo.WeekDuration = weekDur
	if req.Name != nil {
		oldclassInfo.Classname = req.GetName()
	}
	if req.Weeks != nil {
		oldclassInfo.Weeks = req.GetWeeks()
		weekDur := tool.FormatWeeks(tool.ParseWeeks(req.GetWeeks()))
		oldclassInfo.WeekDuration = weekDur
	}
	if req.Credit != nil {
		oldclassInfo.Credit = req.GetCredit()
	}

	oldclassInfo.UpdateID()
	newSc := &model.StudentCourse{
		StuID:           req.GetStuId(),
		ClaID:           oldclassInfo.ID,
		Year:            oldclassInfo.Year,
		Semester:        oldclassInfo.Semester,
		IsManuallyAdded: false,
	}
	newSc.UpdateID()
	err = s.Clu.UpdateClass(ctx, oldclassInfo, newSc, req.GetClassId())
	if err != nil {

		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	return &pb.UpdateClassResponse{
		ClassId: oldclassInfo.ID,
		Msg:     "成功修改",
	}, nil
}
func (s *ClasserService) GetRecycleBinClassInfos(ctx context.Context, req *pb.GetRecycleBinClassRequest) (*pb.GetRecycleBinClassResponse, error) {
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	classInfos, err := s.Clu.GetRecycledClassInfos(ctx)
	if err != nil {
		return &pb.GetRecycleBinClassResponse{}, err
	}
	pbClassInfos := make([]*pb.ClassInfo, 0)
	for _, classInfo := range classInfos {
		pbClassInfos = append(pbClassInfos, HandleClass(classInfo))
	}
	return &pb.GetRecycleBinClassResponse{
		ClassInfos: pbClassInfos,
	}, nil
}
func (s *ClasserService) RecoverClass(ctx context.Context, req *pb.RecoverClassRequest) (*pb.RecoverClassResponse, error) {
	if !tool.CheckSY(req.Semester, req.Year) {
		return &pb.RecoverClassResponse{
			Msg: "恢复课程失败",
		}, errcode.ErrParam
	}
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		StuId:    req.GetStuId(),
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	err := s.Clu.RecoverClassInfo(ctx, req.GetClassId())
	if err != nil {

		return &pb.RecoverClassResponse{
			Msg: "恢复课程失败",
		}, err
	}
	return &pb.RecoverClassResponse{
		Msg: "恢复课程成功",
	}, nil
}
func (s *ClasserService) GetStuIdByJxbId(ctx context.Context, req *pb.GetStuIdByJxbIdRequest) (*pb.GetStuIdByJxbIdResponse, error) {
	stuIds, err := s.Clu.GetStuIdsByJxbId(ctx, req.GetJxbId())
	if err != nil {

		return &pb.GetStuIdByJxbIdResponse{}, errcode.ErrGetStuIdByJxbId
	}
	return &pb.GetStuIdByJxbIdResponse{
		StuId: stuIds,
	}, nil
}
func (s *ClasserService) GetAllClassInfo(ctx context.Context, req *pb.GetAllClassInfoRequest) (*pb.GetAllClassInfoResponse, error) {
	//将stuid,year,semester存入ctx中
	commonInfo := model.CommonInfo{
		Year:     req.GetYear(),
		Semester: req.GetSemester(),
	}
	ctx = model.StoreCommonInfoInCtx(ctx, commonInfo)
	classInfos := s.Clu.GetAllSchoolClassInfosToOtherService(ctx)
	pbClassInfos := make([]*pb.ClassInfo, 0)
	for _, classInfo := range classInfos {
		pbClassInfos = append(pbClassInfos, HandleClass(classInfo))
	}
	return &pb.GetAllClassInfoResponse{
		ClassInfos: pbClassInfos,
	}, nil
}
func HandleClass(info *model.ClassInfo) *pb.ClassInfo {
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
