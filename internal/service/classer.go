package service

import (
	pb "class/api/classer/v1"
	"class/internal/biz"
	"class/internal/errcode"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type ClasserService struct {
	pb.UnimplementedClasserServer
	Clu *biz.ClassUsercase
}

func NewClasserService(clu *biz.ClassUsercase) *ClasserService {
	return &ClasserService{
		Clu: clu,
	}
}

func (s *ClasserService) GetClass(ctx context.Context, req *pb.GetClassRequest) (*pb.GetClassResponse, error) {
	var cli *http.Client //TODO:其他服务给的*http.Client
	pclasses := make([]*pb.Class, 0)

	if !CheckSY(req.Semester, req.Year) || req.GetWeek() <= 0 {
		return &pb.GetClassResponse{}, errcode.ErrParam
	}

	classes, err := s.Clu.GetClasses(ctx, req.GetStuId(), req.GetWeek(), req.GetYear(), req.GetSemester(), cli)
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
func (s *ClasserService) GetOneClass(ctx context.Context, req *pb.GetOneClassRequest) (*pb.GetOneClassResponse, error) {
	pinfos := make([]*pb.ClassInfo, 0)
	if !CheckSY(req.Semester, req.Year) {
		return &pb.GetOneClassResponse{}, errcode.ErrParam
	}
	infos, err := s.Clu.FindClass(ctx, req.GetStuId(), req.GetYear(), req.GetSemester(), req.GetDay(), req.Dur)
	if err != nil {
		return &pb.GetOneClassResponse{}, err
	}
	for _, info := range infos {
		pinfo := HandleClass(info)
		pinfos = append(pinfos, pinfo)
	}
	return &pb.GetOneClassResponse{
		Infos: pinfos,
	}, nil
}
func (s *ClasserService) AddClass(ctx context.Context, req *pb.AddClassRequest) (*pb.AddClassResponse, error) {
	if !CheckSY(req.Semester, req.Year) || req.GetWeeks() <= 0 {
		return &pb.AddClassResponse{}, errcode.ErrParam
	}
	weekDur := FormatWeeks(ParseWeeks(req.Weeks))
	var classInfo = &biz.ClassInfo{
		StuID:           req.GetStuId(),
		Day:             req.GetDay(),
		Teacher:         req.GetTeacher(),
		Where:           req.GetTeacher(),
		ClassWhen:       req.GetDurClass(),
		WeekDuration:    weekDur,
		Classname:       req.GetName(),
		Credit:          req.GetCredit(),
		IsManuallyAdded: true,
		Weeks:           req.GetWeeks(),
		Semester:        req.GetSemester(),
		Year:            req.GetYear(),
	}
	classInfo.UpdateID()
	err := s.Clu.AddClass(ctx, classInfo)
	if err != nil {
		return &pb.AddClassResponse{}, err
	}

	return &pb.AddClassResponse{
		Msg: "成功添加",
	}, nil
}
func (s *ClasserService) DeleteClass(ctx context.Context, req *pb.DeleteClassRequest) (*pb.DeleteClassResponse, error) {
	err := s.Clu.DeleteClass(ctx, req.GetId())
	if err != nil {
		return &pb.DeleteClassResponse{}, err
	}
	return &pb.DeleteClassResponse{
		Msg: "成功删除",
	}, nil
}
func CheckSY(semester, year string) bool {
	var y1, y2 int
	var tag1, tag2 bool
	fmt.Sscanf(year, "%d-%d", &y1, &y2)
	if y2 == y1+1 {
		tag1 = true
	} else {
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
		Add:          info.IsManuallyAdded,
		Weeks:        info.Weeks,
		Id:           info.ID,
		Semester:     info.Semester,
		Year:         info.Year,
		StuId:        info.StuID,
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
