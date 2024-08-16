# Biz层的简单介绍
## biz.go
这个文件主要是提供`wire`依赖注入的`ProviderSet`,以及定义了`Transaction`的接口（其主要是为了优雅地调用gorm中的事务）
## classer.go
1、定义了爬虫的接口，由`pkg/crawler/crawler.go`中的`Crawler`来实现去CCNU教务管理系统中爬取课表
2、定义了`ClassUsercase `，并实现了关于课程的相关方法，以便`service`来调用
## classInfo.go
1、定义了`ClassInfoDBRepo`和`ClassInfoCacheRepo`的接口，由`data`层中`ClassInfoRepo.go`中的相关实例实现，分别控制在mysql和redis中对`ClassInfo`的操作
2、定义了`ClassInfoRepo`，集成了`ClassInfoDBRepo`和`ClassInfoCacheRepo`，使其能够直接控制`ClassInfo`的所有操作
## studentAndCourse.go
1、定义了`StudentAndCourseDBRepo`和`StudentAndCourseCacheRepo`接口，由`data`层中`StudentAndCourse.go`中的相关实例实现，分别控制在mysql和redis中对`StudentCourse`的操作
2、定义了`StudentAndCourseRepo`，集成了`StudentAndCourseDBRepo`和`StudentAndCourseCacheRepo`，使其能够直接控制`StudentCourse`的所有操作
## classRepo.go
便于同时操作`ClassInfo`和`StudentCourse`
## model.go
1、定义`ClassInfo`，其表示课程信息，
2、定义`StudentCourse`，其表示学生与课程之间的对应关系
