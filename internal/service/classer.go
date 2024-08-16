package service

import (
	pb "class/api/classer/v1"
	"class/internal/biz"
	"class/internal/errcode"
	"context"
	v1 "github.com/asynccnu/ccnu-service/api/ccnu_service/v1"
	"sort"
	"strconv"
	"strings"
)

type ClasserService struct {
	pb.UnimplementedClasserServer
	Clu *biz.ClassUsercase
	Cs  v1.CCNUServiceClient
}

func NewClasserService(clu *biz.ClassUsercase, cs v1.CCNUServiceClient) *ClasserService {
	return &ClasserService{
		Clu: clu,
		Cs:  cs,
	}
}

func (s *ClasserService) GetClass(ctx context.Context, req *pb.GetClassRequest) (*pb.GetClassResponse, error) {
	var cookie string
	resp, err := s.Cs.GetCookie(ctx, &v1.GetCookieRequest{
		Userid: req.GetStuId(),
	})
	cookie = resp.Cookie
	pclasses := make([]*pb.Class, 0)

	if !CheckSY(req.Semester, req.Year) || req.GetWeek() <= 0 {
		return &pb.GetClassResponse{}, errcode.ErrParam
	}

	classes, err := s.Clu.GetClasses(ctx, req.GetStuId(), req.GetWeek(), req.GetYear(), req.GetSemester(), cookie)
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
	return &pb.GetClassResponse{
		Classes: pclasses,
	}, nil
}
func (s *ClasserService) AddClass(ctx context.Context, req *pb.AddClassRequest) (*pb.AddClassResponse, error) {
	if !CheckSY(req.Semester, req.Year) || req.GetWeeks() <= 0 {
		return &pb.AddClassResponse{}, errcode.ErrParam
	}
	weekDur := FormatWeeks(ParseWeeks(req.Weeks))
	var classInfo = &biz.ClassInfo{
		Day:          req.GetDay(),
		Teacher:      req.GetTeacher(),
		Where:        req.GetTeacher(),
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
		return &pb.AddClassResponse{}, err
	}

	return &pb.AddClassResponse{
		Id:  classInfo.ID,
		Msg: "成功添加",
	}, nil
}
func (s *ClasserService) DeleteClass(ctx context.Context, req *pb.DeleteClassRequest) (*pb.DeleteClassResponse, error) {
	err := s.Clu.DeleteClass(ctx, req.GetId(), req.GetStuId(), req.GetYear(), req.GetSemester())
	if err != nil {
		return &pb.DeleteClassResponse{}, err
	}
	return &pb.DeleteClassResponse{
		Msg: "成功删除",
	}, nil
}
func (s *ClasserService) UpdateClass(ctx context.Context, req *pb.UpdateClassRequest) (*pb.UpdateClassResponse, error) {
	if !CheckSY(req.Semester, req.GetYear()) || req.GetWeeks() <= 0 {
		return &pb.UpdateClassResponse{}, errcode.ErrParam
	}
	weekDur := FormatWeeks(ParseWeeks(req.GetWeeks()))
	oldclassInfo, err := s.Clu.SearchClass(ctx, req.GetClassId())
	if err != nil {
		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	oldclassInfo.Day = req.GetDay()
	oldclassInfo.Teacher = req.GetTeacher()
	oldclassInfo.Where = req.GetTeacher()
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
		return &pb.UpdateClassResponse{
			Msg: "修改失败",
		}, err
	}
	return &pb.UpdateClassResponse{
		ClassId: oldclassInfo.ID,
		Msg:     "成功修改",
	}, nil
}
func CheckSY(semester, year string) bool {

	var tag1, tag2 bool
	y, err := strconv.Atoi(year)
	if err != nil || y < 2006 {
		tag1 = false
	}
	if semester == "1" || semester == "2" || semester == "3" {
		tag2 = true
	} else {
		tag2 = false
	}
	return tag1 && tag2

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
func ParseWeeks(weeks int64) []int {
	if weeks <= 0 {
		return []int{}
	}
	var weeksList []int
	for i := 1; (1 << (i - 1)) <= weeks; i++ {
		if weeks&(1<<(i-1)) != 0 {
			weeksList = append(weeksList, i)
		}
	}
	return weeksList
}
func FormatWeeks(weeks []int) string {
	if len(weeks) == 0 {
		return ""
	}

	// 对周数集合排序
	sort.Ints(weeks)

	var result strings.Builder
	start := weeks[0]
	end := start
	isSingle := start%2 != 0
	isMixed := false

	// 检查是否是单周、双周还是混合
	for _, week := range weeks {
		if (week%2 == 0) != !isSingle {
			isMixed = true
		}
	}

	// 遍历周数集合，生成格式化字符串
	for i := 1; i < len(weeks); i++ {
		if weeks[i] == end+1 {
			end = weeks[i]
		} else {
			if start == end {
				result.WriteString(strconv.Itoa(start))
			} else {
				result.WriteString(strconv.Itoa(start) + "-" + strconv.Itoa(end))
			}
			result.WriteString(",")
			start = weeks[i]
			end = start
		}
	}

	// 处理最后一段区间
	if start == end {
		result.WriteString(strconv.Itoa(start))
	} else {
		result.WriteString(strconv.Itoa(start) + "-" + strconv.Itoa(end))
	}

	// 添加 "(单)" 或 "(双)" 标识
	if !isMixed {
		if isSingle {
			result.WriteString("周(单)")
		} else {
			result.WriteString("周(双)")
		}
	} else {
		result.WriteString("周")
	}

	return result.String()
}
