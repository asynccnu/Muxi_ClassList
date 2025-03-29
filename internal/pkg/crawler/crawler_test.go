package crawler

import (
	"context"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/model"
	"github.com/asynccnu/Muxi_ClassList/test"
	"testing"
	"time"
)

func TestCrawler_GetClassInfosForUndergraduate(t *testing.T) {
	clog := classLog.NewClogger(test.NewLogger())
	crawler := NewClassCrawler(clog)
	start := time.Now()
	res, err := crawler.GetClassInfosForUndergraduate(context.Background(), model.GetClassInfosForUndergraduateReq{
		StuID:    "testID",
		Year:     "2024",
		Semester: "2",
		Cookie:   "JSESSIONID=ACB2FEEF93678BF837955F63E088D85B",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("一共耗时%v", time.Since(start)))
	t.Log(res)

	start = time.Now()
	res, err = crawler.GetClassInfosForUndergraduate(context.Background(), model.GetClassInfosForUndergraduateReq{
		StuID:    "testID",
		Year:     "2024",
		Semester: "2",
		Cookie:   "JSESSIONID=ACB2FEEF93678BF837955F63E088D85B",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("一共耗时%v", time.Since(start)))
	t.Log(res)
}
