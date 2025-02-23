package model

type ClassBiz struct {
	ID           string
	Day          int64   //星期几
	Teacher      string  //任课教师
	Where        string  //上课地点
	ClassWhen    string  //上课是第几节（如1-2,3-4）
	WeekDuration string  //上课的周数
	Classname    string  //课程名称
	Credit       float64 //学分
	Weeks        int64   //哪些周
	Semester     string  //学期
	Year         string  //学年
}
