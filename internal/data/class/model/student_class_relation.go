package model

import (
	bizmodel "github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"time"
)

const StudentClassRelationDOTableName string = "student_class_relations"

//DO（Data Object）

type StudentClassRelationDO struct {
	StuID           string    `gorm:"type:varchar(20);column:stu_id;not null;uniqueIndex:idx_sc,priority:3" json:"stu_id"`    //学号
	ClaID           string    `gorm:"type:varchar(120);column:cla_id;not null;uniqueIndex:idx_sc,priority:4" json:"cla_id"`   //课程ID
	Year            string    `gorm:"type:varchar(5);column:year;not null;uniqueIndex:idx_sc,priority:1" json:"year"`         //学年
	Semester        string    `gorm:"type:varchar(1);column:semester;not null;uniqueIndex:idx_sc,priority:2" json:"semester"` //学期
	IsManuallyAdded bool      `gorm:"column:is_manually_added;default:false" json:"is_manually_added"`                        //是否为手动添加
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

func (sc *StudentClassRelationDO) TableName() string {
	return StudentClassRelationDOTableName
}

func NewStudentClassRelationDO(stuID, year, semester string, claID string, isManuallyAdded bool) *StudentClassRelationDO {
	createdTime := time.Now()
	return &StudentClassRelationDO{
		StuID:           stuID,
		ClaID:           claID,
		Year:            year,
		Semester:        semester,
		IsManuallyAdded: isManuallyAdded,
		CreatedAt:       createdTime,
		UpdatedAt:       createdTime,
	}
}

func BatchNewStudentClassRelationsDO(stuID, year, semester string, classes []*bizmodel.ClassBiz, isManuallyAdded bool) []*StudentClassRelationDO {
	batch := make([]*StudentClassRelationDO, 0, len(classes))
	for _, class := range classes {
		batch = append(batch, NewStudentClassRelationDO(stuID, year, semester, class.ID, isManuallyAdded))
	}
	return batch
}
