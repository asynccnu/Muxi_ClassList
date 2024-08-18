//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"class/internal/biz"
	"class/internal/client"
	"class/internal/conf"
	"class/internal/data"
	log2 "class/internal/logPrinter"
	"class/internal/pkg"
	"class/internal/pkg/crawler"
	"class/internal/registry"
	"class/internal/server"
	"class/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		pkg.ProviderSet,
		log2.ProviderSet,
		registry.ProviderSet,
		service.ProviderSet,
		client.ProviderSet,
		newApp,
		wire.Bind(new(biz.ClassCrawler), new(*crawler.Crawler)),
		wire.Bind(new(biz.ClassInfoDBRepo), new(*data.ClassInfoDBRepo)),
		wire.Bind(new(biz.ClassInfoCacheRepo), new(*data.ClassInfoCacheRepo)),
		wire.Bind(new(biz.StudentAndCourseDBRepo), new(*data.StudentAndCourseDBRepo)),
		wire.Bind(new(biz.StudentAndCourseCacheRepo), new(*data.StudentAndCourseCacheRepo)),
		wire.Bind(new(service.ClassCtrl), new(*biz.ClassUsercase)),
		wire.Bind(new(service.CCNUServiceProxy), new(*client.CCNUService)),
		wire.Bind(new(biz.ClassRepoProxy), new(*biz.ClassRepo)),
	))
}
