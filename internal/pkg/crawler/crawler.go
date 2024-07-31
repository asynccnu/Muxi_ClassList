package crawler

import (
	"class/internal/biz"
	"class/internal/errcode"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Notice: 爬虫相关
var mp = map[string]string{
	"1": "3",
	"2": "12",
}

type Crawler struct {
	log *log.Helper
}

func NewClassCrawler(logger log.Logger) biz.ClassCrawler {
	return &Crawler{
		log: log.NewHelper(logger),
	}
}

// GetClassInfos 获取课程信息
func (c *Crawler) GetClassInfos(ctx context.Context, client *http.Client, xnm, xqm string) ([]*biz.ClassInfo, error) {

	var reply CrawReply
	tmp1 := GetXNM(xnm)
	tmp2 := GetXQM(xqm)
	formdata := fmt.Sprintf("xnm=%s&xqm=%s&kzlx=ck&xsdm=", tmp1, tmp2)
	var data = strings.NewReader(formdata)
	req, err := http.NewRequest("POST", "https://xk.ccnu.edu.cn/jwglxt/kbcx/xskbcx_cxXsgrkb.html?gnmkdm=N2151", data)
	if err != nil {
		c.log.Errorf("pkg/crawler/crawler.go/GetClassInfos: Error creating request:%v \n", err)
		return nil, errcode.ErrCrawler
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	//req.Header.Set("Cookie", "JSESSIONID=AB63902D520F9FCB33BD3A0E3D9E93DE")
	req.Header.Set("Origin", "https://xk.ccnu.edu.cn")
	req.Header.Set("Referer", "https://xk.ccnu.edu.cn/jwglxt/kbcx/xskbcx_cxXskbcxIndex.html?gnmkdm=N2151&layout=default")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Microsoft Edge";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	resp, err := client.Do(req)
	if err != nil {
		c.log.Errorf("pkg/crawler/crawler.go/GetClassInfos: client.Do(req) failed: %v\n", err)
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		c.log.Errorf("pkg/crawler/crawler.go/GetClassInfos: json.NewDecoder(resp.Body).Decode(&reply):%v\n", err)
		return nil, errcode.ErrCrawler
	}
	infos, err := ToClassInfo(reply, xnm, xqm)
	if err != nil {
		c.log.Errorf("pkg/crawler/crawler.go/GetClassInfos: ToClassInfo(reply, xnm, xqm):%v\n", err)
		return nil, errcode.ErrCrawler
	}
	return infos, nil
}
func ToClassInfo(reply CrawReply, xnm, xqm string) ([]*biz.ClassInfo, error) {
	var infos = make([]*biz.ClassInfo, 0)
	for _, v := range reply.KbList {
		var info = &biz.ClassInfo{}
		//info.ClassID = v.Kch                          //课程ID
		info.StuID = reply.Xsxx.Xh                    //学号
		info.Day, _ = strconv.ParseInt(v.Xqj, 10, 64) //星期几
		info.Teacher = v.Xm                           //教师姓名
		info.Where = v.Cdmc                           //上课地点
		info.ClassWhen = v.Jcs                        //上课是第几节
		info.WeekDuration = v.Zcd                     //上课的周数
		info.Classname = v.Kcmc                       //课程名称
		info.Credit, _ = strconv.ParseFloat(v.Xf, 64) //学分
		info.IsManuallyAdded = false                  //是否为手动添加
		info.Weeks = 0
		info.Semester = xqm //学期
		info.Year = xnm     //学年
		info.UpdateID()     //更新ID
		// 8周,11-15周(单)
		//添加周数
		weeks, err := ParseWeeks(v.Zcd)
		if err != nil {
			return nil, err
		}

		for _, week := range weeks {
			info.AddWeek(week)
		}
		infos = append(infos, info) //添加课程
	}
	return infos, nil
}
func GetXNM(s string) string {
	// 定义正则表达式模式
	re := regexp.MustCompile(`^(\d{4})-\d{4}$`)

	// 查找字符串中与正则表达式模式匹配的部分
	matches := re.FindStringSubmatch(s)

	// 检查是否匹配成功
	if len(matches) > 1 {
		return matches[1] // 第一个捕获组是我们需要的部分
	}
	return ""
}
func GetXQM(s string) string {
	return mp[s]
}
