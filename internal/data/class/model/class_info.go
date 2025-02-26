package model

import (
	"fmt"
	bizmodel "github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"time"
)

const ClassDOTableName string = "classes"

//DO（Data Object）

type ClassDO struct {
	ID        string    `gorm:"type:varchar(150);primaryKey;column:id" json:"id"` //集合了课程信息的字符串，便于标识（课程ID）
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	//ClassId      string  `gorm:"column:class_id" json:"class_id"`           //课程编号
	//JxbId        string  `gorm:"type:varchar(100);column:jxb_id" json:"jxb_id"`                       //教学班ID
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

func (ci *ClassDO) UpdateID() {
	ci.ID = fmt.Sprintf("Class:%s:%s:%s:%d:%s:%s:%s:%d", ci.Classname, ci.Year, ci.Semester, ci.Day, ci.ClassWhen, ci.Teacher, ci.Where, ci.Weeks)
}

func (ci *ClassDO) TableName() string {
	return ClassDOTableName
}

func NewClass(classBiz *bizmodel.ClassBiz) *ClassDO {
	createdTime := time.Now()
	classdo := &ClassDO{
		Day:          classBiz.Day,
		Teacher:      classBiz.Teacher,
		Where:        classBiz.Where,
		ClassWhen:    classBiz.ClassWhen,
		WeekDuration: classBiz.WeekDuration,
		Classname:    classBiz.Classname,
		Credit:       classBiz.Credit,
		Weeks:        classBiz.Weeks,
		Semester:     classBiz.Semester,
		Year:         classBiz.Year,
		CreatedAt:    createdTime,
		UpdatedAt:    createdTime,
	}
	classdo.UpdateID()
	return classdo
}

func BatchNewClasses(classBiz []*bizmodel.ClassBiz) []*ClassDO {
	classdos := make([]*ClassDO, 0, len(classBiz))
	for _, class := range classBiz {
		if class == nil {
			continue
		}
		classdos = append(classdos, NewClass(class))
	}
	return classdos
}
