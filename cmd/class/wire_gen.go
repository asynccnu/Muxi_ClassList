// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/asynccnu/Muxi_ClassList/internal/biz"
	"github.com/asynccnu/Muxi_ClassList/internal/client"
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/asynccnu/Muxi_ClassList/internal/data"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/cache"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/repo"
	"github.com/asynccnu/Muxi_ClassList/internal/data/jxb"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/crawler"
	"github.com/asynccnu/Muxi_ClassList/internal/registry"
	"github.com/asynccnu/Muxi_ClassList/internal/server"
	"github.com/asynccnu/Muxi_ClassList/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, confData *conf.Data, confRegistry *conf.Registry, schoolDay *conf.SchoolDay, logger log.Logger) (*kratos.App, func(), error) {
	redisClient := data.NewRedisDB(confData)
	cacheCache := cache.NewCache(redisClient)
	db := data.NewDB(confData)
	classRepo := repo.NewClassRepo(cacheCache, cacheCache, db)
	crawlerCrawler := crawler.NewClassCrawler()
	jxbDBRepo := jxb.NewJxbDBRepo(db)
	etcdRegistry := registry.NewRegistrarServer(confRegistry, logger)
	userServiceClient, err := client.NewClient(etcdRegistry, logger)
	if err != nil {
		return nil, nil, err
	}
	ccnuService := client.NewCCNUService(userServiceClient)
	classUsecase := biz.NewClassUsecase(classRepo, classRepo, classRepo, classRepo, crawlerCrawler, jxbDBRepo, ccnuService)
	classerService := service.NewClasserService(classUsecase, schoolDay)
	grpcServer := server.NewGRPCServer(confServer, classerService, logger)
	app := newApp(logger, grpcServer, etcdRegistry)
	return app, func() {
	}, nil
}
