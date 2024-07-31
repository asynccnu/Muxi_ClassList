package biz

import (
	"class/internal/errcode"
	log2 "class/internal/log"
	"context"
	"errors"
	"net/http"
)

// ClassRepo 数据持久化接口
type ClassRepo interface {
	SaveClassInfo(ctx context.Context, cla []*ClassInfo) error
	AddClassInfo(ctx context.Context, cla *ClassInfo) error
	GetSpecificClassInfo(ctx context.Context, id string, xnm, xqm string, day int64, dur string) ([]*ClassInfo, error)
	GetClasses(ctx context.Context, id string, xnm, xqm string) ([]*ClassInfo, error)
	DeleteClass(ctx context.Context, id string) error
}

// ClassCrawler 课程爬虫接口
type ClassCrawler interface {
	GetClassInfos(ctx context.Context, client *http.Client, xnm, xqm string) ([]*ClassInfo, error)
}

type ClassUsercase struct {
	Repo    ClassRepo
	Crawler ClassCrawler
	log     log2.LogerPrinter
}

func NewClassUsercase(repo ClassRepo, crawler ClassCrawler, log log2.LogerPrinter) *ClassUsercase {
	return &ClassUsercase{
		Repo:    repo,
		Crawler: crawler,
		log:     log,
	}
}

func (cluc *ClassUsercase) GetClasses(ctx context.Context, id string, week int64, xnm, xqm string, client *http.Client) ([]*Class, error) {
	var classInfos = make([]*ClassInfo, 0)
	var classes = make([]*Class, 0)
	var err error
	classInfos, err = cluc.Repo.GetClasses(ctx, id, xnm, xqm)
	if err != nil {
		//如果数据库中没有就去爬
		if errors.Is(err, errcode.ErrClassNotFound) {

			classInfos, err = cluc.Crawler.GetClassInfos(ctx, client, xnm, xqm)

			if err != nil {
				cluc.log.FuncError(cluc.Crawler.GetClassInfos, err)
				return nil, err
			}
			go func() {
				err := cluc.Repo.SaveClassInfo(ctx, classInfos)
				if err != nil {
					cluc.log.FuncError(cluc.Repo.SaveClassInfo, err)
				}
			}()
		}

		return nil, err
	}

	for _, classInfo := range classInfos {
		thisWeek := classInfo.SearchWeek(week)
		class := &Class{
			Info:     classInfo,
			ThisWeek: thisWeek,
		}
		classes = append(classes, class)
	}

	return classes, nil
}
func (cluc *ClassUsercase) FindClass(ctx context.Context, id string, xnm, xqm string, day int64, dur string) ([]*ClassInfo, error) {

	classInfos, err := cluc.Repo.GetSpecificClassInfo(ctx, id, xnm, xqm, day, dur)
	if err != nil {
		cluc.log.FuncError(cluc.Repo.GetSpecificClassInfo, err)
		return nil, err
	}
	return classInfos, nil
}
func (cluc *ClassUsercase) AddClass(ctx context.Context, info *ClassInfo) error {
	err := cluc.Repo.AddClassInfo(ctx, info)
	if err != nil {
		cluc.log.FuncError(cluc.Repo.AddClassInfo, err)
		return err
	}
	return nil
}
func (cluc *ClassUsercase) DeleteClass(ctx context.Context, id string) error {
	err := cluc.Repo.DeleteClass(ctx, id)
	if err != nil {
		cluc.log.FuncError(cluc.Repo.DeleteClass, err)
		return err
	}
	return nil
}
