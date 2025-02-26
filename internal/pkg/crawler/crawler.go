package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/errcode"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"net/http"
	"strconv"
	"strings"
)

// Notice: 爬虫相关
var mp = map[string]string{
	"1": "3",
	"2": "12",
	"3": "16",
}

type Crawler struct{}

func NewClassCrawler() *Crawler {
	return &Crawler{}
}

// GetClassInfoForGraduateStudent 获取研究生课程信息
func (c *Crawler) GetClassInfoForGraduateStudent(ctx context.Context, year, semester, cookie string) ([]*model.ClassBiz, []string, error) {
	var (
		reply CrawReply2
		xnm   = year
		xqm   = semester
		yn    = tool.CheckSY(xqm, xnm)
	)

	if !yn {
		return nil, nil, errcode.ErrParam
	}
	client := &http.Client{}
	param := fmt.Sprintf("xnm=%s&xqm=%s", GetXNM(xnm), GetXQM(xqm))
	var data = strings.NewReader(param)
	req, err := http.NewRequest("POST", "https://grd.ccnu.edu.cn/yjsxt/kbcx/xskbcx_cxXsKb.html?gnmkdm=N2151", data)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Origin", "https://grd.ccnu.edu.cn")
	req.Header.Set("Referer", "https://grd.ccnu.edu.cn/yjsxt/kbcx/xskbcx_cxXskbcxIndex.html?gnmkdm=N2151&layout=default")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36 Edg/127.0.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Not)A;Brand";v="99", "Microsoft Edge";v="127", "Chromium";v="127"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	infos, jxbIDs, err := ToClassInfo2(reply, xnm, xqm)
	if err != nil {
		return nil, nil, nil
	}
	return infos, jxbIDs, nil
}

// GetClassInfosForUndergraduate  获取本科生课程信息
func (c *Crawler) GetClassInfosForUndergraduate(ctx context.Context, year, semester, cookie string) ([]*model.ClassBiz, []string, error) {
	var (
		reply CrawReply1
		xnm   = year
		xqm   = semester
		yn    = tool.CheckSY(xqm, xnm)
	)

	if !yn {
		return nil, nil, errcode.ErrParam
	}

	client := &http.Client{}
	formdata := fmt.Sprintf("xnm=%s&xqm=%s&kzlx=ck&xsdm=", GetXNM(xnm), GetXQM(xqm))
	var data = strings.NewReader(formdata)
	req, err := http.NewRequest("POST", "https://xk.ccnu.edu.cn/jwglxt/kbcx/xskbcx_cxXsgrkb.html?gnmkdm=N2151", data)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	req.Header.Set("Cookie", cookie) //设置cookie
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
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
		return nil, nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()
	//fmt.Println(resp)
	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	infos, jxbIDs, err := ToClassInfo1(reply, xnm, xqm)
	if err != nil {
		return nil, nil, errcode.ErrCrawler
	}
	return infos, jxbIDs, nil
}

// ToClassInfo1 处理本科生
func ToClassInfo1(reply CrawReply1, xnm, xqm string) ([]*model.ClassBiz, []string, error) {
	var infos []*model.ClassBiz
	var jxbIDs []string
	for _, v := range reply.KbList {
		//这个字段为1代表这个课程是已经选上的
		//所以防止把没选上的课程给包含进来
		if v.Sxbj != "1" {
			continue
		}

		//课程信息
		info := new(model.ClassBiz)
		//var Sc = &biz.StudentCourse{}
		//info.ClassId = v.Kch //课程编号
		//info.StuID = reply.Xsxx.Xh                    //学号
		info.Day, _ = strconv.ParseInt(v.Xqj, 10, 64) //星期几
		info.Teacher = v.Xm                           //教师姓名
		info.Where = v.Cdmc                           //上课地点
		info.ClassWhen = v.Jcs                        //上课是第几节
		info.WeekDuration = v.Zcd                     //上课的周数
		info.Classname = v.Kcmc                       //课程名称
		info.Credit, _ = strconv.ParseFloat(v.Xf, 64) //学分
		info.Semester = xqm                           //学期
		info.Year = xnm                               //学年
		//添加周数
		info.Weeks, _ = strconv.ParseInt(v.Oldzc, 10, 64)

		jxbIDs = append(jxbIDs, v.JxbID)

		infos = append(infos, info) //添加课程
	}
	return infos, jxbIDs, nil
}

// ToClassInfo2 处理研究生
func ToClassInfo2(reply CrawReply2, xnm, xqm string) ([]*model.ClassBiz, []string, error) {
	var infos []*model.ClassBiz
	var jxbIDs []string
	for _, v := range reply.KbList {
		//课程信息
		info := new(model.ClassBiz)
		//var Sc = &biz.StudentCourse{}
		//info.ClassId = v.Kch //课程编号
		//info.StuID = reply.Xsxx.Xh                    //学号
		info.Day, _ = strconv.ParseInt(v.Xqj, 10, 64) //星期几
		info.Teacher = v.Xm                           //教师姓名
		info.Where = v.Cdmc                           //上课地点
		info.ClassWhen = v.Jcs                        //上课是第几节
		info.WeekDuration = v.Zcd                     //上课的周数
		info.Classname = v.Kcmc                       //课程名称
		info.Credit, _ = strconv.ParseFloat(v.Xf, 64) //学分
		info.Semester = xqm                           //学期
		info.Year = xnm                               //学年
		//添加周数
		info.Weeks, _ = strconv.ParseInt(v.Oldzc, 10, 64)

		jxbIDs = append(jxbIDs, v.JxbID) //添加教学班ID

		infos = append(infos, info) //添加课程
	}
	return infos, jxbIDs, nil
}
func GetXNM(s string) string {
	// // 定义正则表达式模式
	// re := regexp.MustCompile(`^(\d{4})-\d{4}$`)

	// // 查找字符串中与正则表达式模式匹配的部分
	// matches := re.FindStringSubmatch(s)

	// // 检查是否匹配成功
	// if len(matches) > 1 {
	// 	return matches[1] // 第一个捕获组是我们需要的部分
	// }
	return s
}
func GetXQM(s string) string {
	return mp[s]
}
