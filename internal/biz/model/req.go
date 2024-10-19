package model

import "context"

const (
	COMMONINFO = "commoninfo"
)

type CommonInfo struct {
	StuId    string
	Year     string
	Semester string
}

type GetClassInfosForUndergraduateReq struct {
	Cookie string
}
type GetClassInfoForGraduateStudentReq struct {
	Cookie string
}
type SaveClassReq struct {
	ClassInfos []*ClassInfo
	Scs        []*StudentCourse
}
type GetSpecificClassInfoReq struct {
	ClassId string
}
type AddClassReq struct {
	ClassInfo *ClassInfo
	Sc        *StudentCourse
}
type DeleteClassReq struct {
	ClassId string
}
type RecoverClassFromRecycleBinReq struct {
	ClassId string
}
type UpdateClassReq struct {
	NewClassInfo *ClassInfo
	NewSc        *StudentCourse
	OldClassId   string
}
type CheckSCIdsExistReq struct {
	ClassId string
}
type CheckClassIdIsInRecycledBinReq struct {
	ClassId string
}

func StoreCommonInfoInCtx(ctx context.Context, info CommonInfo) context.Context {
	return context.WithValue(ctx, COMMONINFO, info)
}
func GetCommonInfoFromCtx(ctx context.Context) CommonInfo {
	return ctx.Value(COMMONINFO).(CommonInfo)
}
