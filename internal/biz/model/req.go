package model

type GetClassInfosForUndergraduateReq struct {
	Cookie string
	Xnm    string
	Xqm    string
}
type GetClassInfoForGraduateStudentReq struct {
	Cookie string
	Xnm    string
	Xqm    string
}
type SaveClassReq struct {
	StuId      string
	Xnm        string
	Xqm        string
	ClassInfos []*ClassInfo
	Scs        []*StudentCourse
}
type GetAllClassesReq struct {
	StuId string
	Xnm   string
	Xqm   string
}
type GetSpecificClassInfoReq struct {
	ClassId string
}
type AddClassReq struct {
	ClassInfo *ClassInfo
	Sc        *StudentCourse
	Xnm       string
	Xqm       string
}
type DeleteClassReq struct {
	ClassId string
	StuId   string
	Xnm     string
	Xqm     string
}
type GetRecycledIdsReq struct {
	StuId string
	Xnm   string
	Xqm   string
}
type RecoverClassFromRecycleBinReq struct {
	StuId   string
	Xnm     string
	Xqm     string
	ClassId string
}
type UpdateClassReq struct {
	NewClassInfo *ClassInfo
	NewSc        *StudentCourse
	StuId        string
	OldClassId   string
	Xnm          string
	Xqm          string
}
type CheckSCIdsExistReq struct {
	StuId   string
	ClassId string
	Xnm     string
	Xqm     string
}
type GetAllSchoolClassInfosReq struct {
	Xnm string
	Xqm string
}
type CheckClassIdIsInRecycledBinReq struct {
	StuId   string
	Xnm     string
	Xqm     string
	ClassId string
}
