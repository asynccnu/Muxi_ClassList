package pkg

import (
	"class/internal/pkg/crawler"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(crawler.NewClassCrawler)
