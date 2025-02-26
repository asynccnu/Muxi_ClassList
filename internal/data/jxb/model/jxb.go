package model

const JxbTableName string = "jxbs"

// Jxb 用来存取教学班
type Jxb struct {
	JxbId string `gorm:"type:varchar(100);column:jxb_id;uniqueIndex:idx_jxb,priority:1" json:"jxb_id"` // 教学班ID
	StuId string `gorm:"type:varchar(20);column:stu_id;uniqueIndex:idx_jxb,priority:2" json:"stu_id"`  // 学号
}

func (j *Jxb) TableName() string {
	return JxbTableName
}
