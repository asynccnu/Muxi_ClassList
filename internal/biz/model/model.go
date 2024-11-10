package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

const (
	ClassInfoTableName     string = "class_info"
	StudentCourseTableName string = "student_course"
	JxbTableName           string = "jxb"
)

type Class struct {
	Info     *ClassInfo //课程信息
	ThisWeek bool       //是否是本周
}
type ClassInfo struct {
	ID        string    `gorm:"type:varchar(150);primaryKey;column:id" json:"id"` //集合了课程信息的字符串，便于标识（课程ID）
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	//ClassId      string  `gorm:"column:class_id" json:"class_id"`           //课程编号
	JxbId        string  `gorm:"type:varchar(100);column:jxb_id" json:"jxb_id"`                       //教学班ID
	Day          int64   `gorm:"column:day;not null" json:"day"`                                      //星期几
	Teacher      string  `gorm:"type:varchar(50);;column:teacher;not null" json:"teacher"`            //任课教师
	Where        string  `gorm:"type:varchar(50);column:where;not null" json:"where"`                 //上课地点
	ClassWhen    string  `gorm:"type:varchar(10);column:class_when;not null" json:"class_when"`       //上课是第几节（如1-2,3-4）
	WeekDuration string  `gorm:"type:varchar(20);column:week_duration;not null" json:"week_duration"` //上课的周数
	Classname    string  `gorm:"type:varchar(20);column:class_name;not null" json:"classname"`        //课程名称
	Credit       float64 `gorm:"column:credit;default:1.0" json:"credit"`                             //学分
	Weeks        int64   `gorm:"column:weeks;not null" json:"weeks"`                                  //哪些周
	Semester     string  `gorm:"type:varchar(1);column:semester;not null" json:"semester"`            //学期
	Year         string  `gorm:"type:varchar(5);column:year;not null" json:"year"`                    //学年
}
type StudentCourse struct {
	ID              string    `gorm:"primaryKey;column:id" json:"id"`
	StuID           string    `gorm:"type:varchar(20);column:stu_id;not null;index" json:"stu_id"`                        //学号
	ClaID           string    `gorm:"type:varchar(120);column:cla_id;not null;index" json:"cla_id"`                       //课程ID
	Year            string    `gorm:"type:varchar(5);column:year;not null;index:idx_time,priority:1" json:"year"`         //学年
	Semester        string    `gorm:"type:varchar(1);column:semester;not null;index:idx_time,priority:2" json:"semester"` //学期
	IsManuallyAdded bool      `gorm:"column:is_manually_added;default:false" json:"is_manually_added"`                    //是否为手动添加
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

// Jxb 用来存取教学班
type Jxb struct {
	JxbId string `gorm:"type:varchar(100);column:jxb_id;index" json:"jxb_id"` // 教学班ID
	StuId string `gorm:"type:varchar(20);column:stu_id;index" json:"stu_id"`  // 学号
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

func (j *Jxb) TableName() string {
	return JxbTableName
}

func (ci *ClassInfo) AddWeek(week int64) {
	ci.Weeks |= 1<<week - 1
}

func (ci *ClassInfo) SearchWeek(week int64) bool {
	return (ci.Weeks & (1<<week - 1)) != 0
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
