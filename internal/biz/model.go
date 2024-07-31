package biz

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

const (
	ClassInfoTableName string = "class_info"
)

type Class struct {
	Info     *ClassInfo //课程信息
	ThisWeek bool       //是否是本周
}
type ClassInfo struct {
	ID        string `gorm:"primaryKey;column:id" json:"id"` //集合了课程信息的字符串，便于标识
	CreatedAt time.Time
	UpdatedAt time.Time
	//ClassID         string  `gorm:"column:class_id" json:"class_id"`               //课程ID
	StuID           string  `gorm:"column:stu_id" json:"stu_id"`                   //学号
	Day             int64   `gorm:"column:day" json:"day"`                         //星期几
	Teacher         string  `gorm:"column:teacher" json:"teacher"`                 //任课教师
	Where           string  `gorm:"column:where" json:"where"`                     //上课地点
	ClassWhen       string  `gorm:"column:class_when" json:"class_when"`           //上课是第几节（如1-2,3-4）
	WeekDuration    string  `gorm:"column:week_duration" json:"week_duration"`     //上课的周数
	Classname       string  `gorm:"column:class_name" json:"classname"`            //课程名称
	Credit          float64 `gorm:"column:credit" json:"credit"`                   //学分
	IsManuallyAdded bool    `gorm:"column:isManuallyAdded" json:"IsManuallyAdded"` //是否为手动添加
	Weeks           int64   `gorm:"column:weeks" json:"weeks"`                     //哪些周
	Semester        string  `gorm:"column:semester" json:"semester"`               //学期
	Year            string  `gorm:"column:year" json:"year"`                       //学年
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

func (ci *ClassInfo) AddWeek(week int64) {
	ci.Weeks |= 1<<week - 1
}
func (ci *ClassInfo) SearchWeek(week int64) bool {
	return (ci.Weeks & (1<<week - 1)) == 1
}
func (ci *ClassInfo) GetKey() string {
	return fmt.Sprintf("class:%s:%s:%s:%d:%s:%s", ci.StuID, ci.Year, ci.Semester, ci.Day, ci.ClassWhen, ci.ID)
}
func (ci *ClassInfo) UpdateID() {
	ci.ID = fmt.Sprintf("%s:%s:%s:%s:%d:%s:%s:%s:%d", ci.StuID, ci.Classname, ci.Year, ci.Semester, ci.Day, ci.ClassWhen, ci.Teacher, ci.Where, ci.Weeks)
}
