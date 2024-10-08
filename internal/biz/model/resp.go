package model

type GetClassInfosForUndergraduateResp struct {
	ClassInfos     []*ClassInfo
	StudentCourses []*StudentCourse
}
type GetClassInfoForGraduateStudentResp struct {
	ClassInfos     []*ClassInfo
	StudentCourses []*StudentCourse
}
type GetAllClassesResp struct {
	ClassInfos []*ClassInfo
}
type GetSpecificClassInfoResp struct {
	ClassInfo *ClassInfo
}
type GetRecycledIdsResp struct {
	Ids []string
}
type GetAllSchoolClassInfosResp struct {
	ClassInfos []*ClassInfo
}
