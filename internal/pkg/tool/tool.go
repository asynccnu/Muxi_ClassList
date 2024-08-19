package tool

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

func CheckSY(semester, year string) bool {

	var tag1, tag2 bool
	y, err := strconv.Atoi(year)
	currentYear := time.Now().Year()
	if err != nil || y < 2006 || y >= currentYear+2 { //年份小于2006或者年份大于后年的不予处理
		tag1 = false
	} else {
		tag1 = true
	}
	if semester == "1" || semester == "2" || semester == "3" {
		tag2 = true
	} else {
		tag2 = false
	}
	return tag1 && tag2

}
func ParseWeeks(weeks int64) []int {
	if weeks <= 0 {
		return []int{}
	}
	var weeksList []int
	for i := 1; (1 << (i - 1)) <= weeks; i++ {
		if weeks&(1<<(i-1)) != 0 {
			weeksList = append(weeksList, i)
		}
	}
	return weeksList
}
func FormatWeeks(weeks []int) string {
	if len(weeks) == 0 {
		return ""
	}

	// 对周数集合排序
	sort.Ints(weeks)

	var result strings.Builder
	start := weeks[0]
	end := start
	isSingle := start%2 != 0
	isMixed := false

	// 检查是否是单周、双周还是混合
	for _, week := range weeks {
		if (week%2 == 0) != !isSingle {
			isMixed = true
		}
	}

	// 遍历周数集合，生成格式化字符串
	for i := 1; i < len(weeks); i++ {
		if weeks[i] == end+1 {
			end = weeks[i]
		} else {
			if start == end {
				result.WriteString(strconv.Itoa(start))
			} else {
				result.WriteString(strconv.Itoa(start) + "-" + strconv.Itoa(end))
			}
			result.WriteString(",")
			start = weeks[i]
			end = start
		}
	}

	// 处理最后一段区间
	if start == end {
		result.WriteString(strconv.Itoa(start))
	} else {
		result.WriteString(strconv.Itoa(start) + "-" + strconv.Itoa(end))
	}

	// 添加 "(单)" 或 "(双)" 标识
	if !isMixed {
		if isSingle {
			result.WriteString("周(单)")
		} else {
			result.WriteString("周(双)")
		}
	} else {
		result.WriteString("周")
	}

	return result.String()
}
func CheckIfThisWeek(xnm, xqm string) bool {
	y, _ := strconv.Atoi(xnm)
	s, _ := strconv.Atoi(xqm)
	currentYear := time.Now().Year()
	currentMonth := time.Now().Month()
	//currentYear := 2023
	//currentMonth := 10
	if currentMonth >= 9 {
		return (y == currentYear) && (s == 1)
	}
	if currentMonth <= 1 {
		return (y == currentYear-1) && (s == 1)
	}
	if currentMonth >= 2 && currentMonth <= 6 {
		return (y == currentYear-1) && (s == 2)
	}
	if currentMonth >= 7 && currentMonth <= 8 {
		return (y == currentYear-1) && (s == 3)
	}
	return false

}
