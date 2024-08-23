package pkg

import (
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/crawler"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(crawler.NewClassCrawler)
