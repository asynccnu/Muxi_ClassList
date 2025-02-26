package biz

import (
	"context"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
)

type Student interface {
	GetClass(ctx context.Context, stuID, year, semester, cookie string, craw ClassCrawler) ([]*model.ClassBiz, []string, error)
}
type Undergraduate struct{}

func (u *Undergraduate) GetClass(ctx context.Context, stuID, year, semester, cookie string, craw ClassCrawler) ([]*model.ClassBiz, []string, error) {
	return craw.GetClassInfosForUndergraduate(ctx, year, semester, cookie)

}

type GraduateStudent struct{}

func (g *GraduateStudent) GetClass(ctx context.Context, stuID, year, semester, cookie string, craw ClassCrawler) ([]*model.ClassBiz, []string, error) {
	return craw.GetClassInfoForGraduateStudent(ctx, year, semester, cookie)
}
