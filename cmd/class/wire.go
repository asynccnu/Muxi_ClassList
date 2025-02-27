//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/asynccnu/Muxi_ClassList/internal/biz"
	"github.com/asynccnu/Muxi_ClassList/internal/client"
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/asynccnu/Muxi_ClassList/internal/data"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/cache"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/repo"
	"github.com/asynccnu/Muxi_ClassList/internal/data/jxb"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/crawler"
	"github.com/asynccnu/Muxi_ClassList/internal/registry"
	"github.com/asynccnu/Muxi_ClassList/internal/server"
	"github.com/asynccnu/Muxi_ClassList/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Registry, *conf.SchoolDay, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		pkg.ProviderSet,
		registry.ProviderSet,
		service.ProviderSet,
		client.ProviderSet,
		newApp,
		wire.Bind(new(biz.ClassCrawler), new(*crawler.Crawler)),
		wire.Bind(new(biz.ClassStorage), new(*repo.ClassRepo)),
		wire.Bind(new(biz.ClassRecycleBinManager), new(*repo.ClassRepo)),
		wire.Bind(new(biz.ManualClassManager), new(*repo.ClassRepo)),
		wire.Bind(new(biz.SchoolClassExplorer), new(*repo.ClassRepo)),
		wire.Bind(new(biz.JxbRepo), new(*jxb.JxbDBRepo)),
		wire.Bind(new(biz.CCNUServiceProxy), new(*client.CCNUService)),
		wire.Bind(new(repo.ClassCache), new(*cache.Cache)),
		wire.Bind(new(repo.RecycleBinCache), new(*cache.Cache)),
	))
}
