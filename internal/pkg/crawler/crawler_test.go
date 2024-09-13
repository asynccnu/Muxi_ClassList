package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

var ctx = context.Background()

func TestCrawler_GetClassInfos(t *testing.T) {
	var cookie = "JSESSIONID=E48CAEEB7D2EA3CF0ABE01546CCCDE13"
	type args struct {
		ctx    context.Context
		cookie string
		xnm    string
		xqm    string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test1:2023/1", args{ctx, cookie, "2023", "1"}},
		{"Test2:2023/2", args{ctx, cookie, "2023", "2"}},
		{"Test2:2023/3", args{ctx, cookie, "2023", "3"}},
		{"Test3:2024/1", args{ctx, cookie, "2024", "1"}},
		{"Test4:2024/2", args{ctx, cookie, "2024", "2"}},
		{"Test4:2026/1", args{ctx, cookie, "2026", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Crawler{
				//log: tt.fields.log,
				client: &http.Client{},
			}
			got, got1, err := c.GetClassInfosForUndergraduate(tt.args.ctx, tt.args.cookie, tt.args.xnm, tt.args.xqm)

			fmt.Println("-----------------------------------------------------------------------------")
			fmt.Println(tt.name + ":")
			if err != nil {
				t.Log(err)
				return
			}
			jsonStr1, _ := json.MarshalIndent(got, "", "  ")
			fmt.Println(string(jsonStr1))
			jsonStr2, _ := json.MarshalIndent(got1, "", "  ")
			fmt.Println(string(jsonStr2))
		})
	}
}

func TestCrawler_GetClassInfoForGraduateStudent(t *testing.T) {
	cookie := "JSESSIONID=7160BE00B3CB95BDE5C793A889D15189; route=97b58dd3002fa63e4590a6f4997064a6"
	type args struct {
		ctx    context.Context
		cookie string
		xnm    string
		xqm    string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"test1", args{ctx, cookie, "2024", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Crawler{
				//log: tt.fields.log,
				client: &http.Client{},
			}
			got, got1, err := c.GetClassInfoForGraduateStudent(tt.args.ctx, tt.args.cookie, tt.args.xnm, tt.args.xqm)
			fmt.Println("-----------------------------------------------------------------------------")
			fmt.Println(tt.name + ":")
			if err != nil {
				t.Log(err)
				return
			}
			jsonStr1, _ := json.MarshalIndent(got, "", "  ")
			fmt.Println(string(jsonStr1))
			jsonStr2, _ := json.MarshalIndent(got1, "", "  ")
			fmt.Println(string(jsonStr2))
		})
	}
}
