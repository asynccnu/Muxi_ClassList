//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"class/internal/biz"
	"class/internal/conf"
	"class/internal/data"
	log2 "class/internal/log"
	"class/internal/pkg"
	"class/internal/registry"
	"class/internal/server"
	"class/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, pkg.ProviderSet, log2.ProviderSet, registry.ProviderSet, service.ProviderSet, newApp))
}
