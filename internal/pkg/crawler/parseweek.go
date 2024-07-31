package crawler

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseWeeks 解析输入字符串，并返回包含解析结果的 int64 切片
func ParseWeeks(input string) ([]int64, error) {
	var result []int64

	// 分割逗号分隔的范围
	parts := strings.Split(input, ",")

	for _, part := range parts {
		// 匹配单周和双周
		singleWeek := false
		doubleWeek := false

		if strings.Contains(part, "(单)") {
			singleWeek = true
			part = strings.Replace(part, "(单)", "", -1)
		}
		if strings.Contains(part, "(双)") {
			doubleWeek = true
			part = strings.Replace(part, "(双)", "", -1)
		}

		// 匹配单个周数或范围
		re := regexp.MustCompile(`(\d+)-(\d+)|(\d+)`)
		matches := re.FindAllStringSubmatch(part, -1)

		for _, match := range matches {
			if match[1] != "" && match[2] != "" { // 范围匹配
				start, err := strconv.ParseInt(match[1], 10, 64)
				if err != nil {
					return nil, err
				}
				end, err := strconv.ParseInt(match[2], 10, 64)
				if err != nil {
					return nil, err
				}

				for i := start; i <= end; i++ {
					if singleWeek && i%2 == 0 {
						continue
					}
					if doubleWeek && i%2 != 0 {
						continue
					}
					result = append(result, i)
				}
			} else if match[3] != "" { // 单个周数匹配
				week, err := strconv.ParseInt(match[3], 10, 64)
				if err != nil {
					return nil, err
				}
				if singleWeek && week%2 == 0 {
					continue
				}
				if doubleWeek && week%2 != 0 {
					continue
				}
				result = append(result, week)
			}
		}
	}

	// 去重并排序
	result = removeDuplicates(result)

	return result, nil
}

// removeDuplicates 去除切片中的重复元素
func removeDuplicates(elements []int64) []int64 {
	encountered := map[int64]bool{}
	result := []int64{}

	for _, v := range elements {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result
}
