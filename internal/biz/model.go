package biz

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

const (
	ClassInfoTableName     string = "class_info"
	StudentCourseTableName string = "student_course"
)

type Class struct {
	Info     *ClassInfo //课程信息
	StuID    string     //学号
	ThisWeek bool       //是否是本周
}
type ClassInfo struct {
	ID        string `gorm:"primaryKey;column:id" json:"id"` //集合了课程信息的字符串，便于标识（课程ID）
	CreatedAt time.Time
	UpdatedAt time.Time
	//ClassId      string  `gorm:"column:class_id" json:"class_id"`           //课程编号
	Day          int64   `gorm:"column:day" json:"day"`                     //星期几
	Teacher      string  `gorm:"column:teacher" json:"teacher"`             //任课教师
	Where        string  `gorm:"column:where" json:"where"`                 //上课地点
	ClassWhen    string  `gorm:"column:class_when" json:"class_when"`       //上课是第几节（如1-2,3-4）
	WeekDuration string  `gorm:"column:week_duration" json:"week_duration"` //上课的周数
	Classname    string  `gorm:"column:class_name" json:"classname"`        //课程名称
	Credit       float64 `gorm:"column:credit" json:"credit"`               //学分
	Weeks        int64   `gorm:"column:weeks" json:"weeks"`                 //哪些周
	Semester     string  `gorm:"column:semester" json:"semester"`           //学期
	Year         string  `gorm:"column:year" json:"year"`                   //学年
}
type StudentCourse struct {
	ID              string `gorm:"primaryKey;column:id" json:"id"`
	StuID           string `gorm:"column:stu_id" json:"stu_id"`                       //学号
	ClaID           string `gorm:"column:cla_id" json:"cla_id"`                       //课程ID
	Year            string `gorm:"column:year" json:"year"`                           //学年
	Semester        string `gorm:"column:semester" json:"semester"`                   //学期
	IsManuallyAdded bool   `gorm:"column:is_manually_added" json:"is_manually_added"` //是否为手动添加
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (ci *ClassInfo) TableName() string {
	return ClassInfoTableName
}
func (ci *ClassInfo) BeforeCreate(tx *gorm.DB) (err error) {
	ci.CreatedAt = time.Now()
	ci.UpdatedAt = time.Now()
	return
}
func (ci *ClassInfo) BeforeUpdate(tx *gorm.DB) (err error) {
	ci.UpdatedAt = time.Now()
	return
}
func (sc *StudentCourse) TableName() string {
	return StudentCourseTableName
}
func (sc *StudentCourse) BeforeCreate(tx *gorm.DB) (err error) {
	sc.CreatedAt = time.Now()
	sc.UpdatedAt = time.Now()
	return
}
func (sc *StudentCourse) BeforeUpdate(tx *gorm.DB) (err error) {
	sc.UpdatedAt = time.Now()
	return
}

func (ci *ClassInfo) AddWeek(week int64) {
	ci.Weeks |= 1<<week - 1
}
func (ci *ClassInfo) SearchWeek(week int64) bool {
	return (ci.Weeks & (1<<week - 1)) == 1
}
func (ci *ClassInfo) GetStartAndEndFromClassWhen() (int64, int64) {
	var start, end int64
	if _, err := fmt.Sscanf(ci.ClassWhen, "%d-%d", &start, &end); err == nil {
		return start, end
	}
	return -1, -1
}
func (ci *ClassInfo) UpdateID() {
	ci.ID = fmt.Sprintf("Class:%s:%s:%s:%d:%s:%s:%s:%d", ci.Classname, ci.Year, ci.Semester, ci.Day, ci.ClassWhen, ci.Teacher, ci.Where, ci.Weeks)
}
func (sc *StudentCourse) UpdateID() {
	sc.ID = fmt.Sprintf("StuAndCla:%s:%s:%s:%s", sc.StuID, sc.ClaID, sc.Year, sc.Semester)
}
func GenerateSCID(stuId, classId, xnm, xqm string) string {
	return fmt.Sprintf("StuAndCla:%s:%s:%s:%s", stuId, classId, xnm, xqm)
}
