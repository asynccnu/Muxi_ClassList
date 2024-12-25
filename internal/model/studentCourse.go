package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

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

func (sc *StudentCourse) UpdateID() {
	sc.ID = fmt.Sprintf("StuAndCla:%s:%s:%s:%s", sc.StuID, sc.ClaID, sc.Year, sc.Semester)
}
