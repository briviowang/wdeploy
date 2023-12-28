package runtime

import (
	"strings"
)

// PHPClass 表示一个php类
type PHPClass struct {
	ClassName    string
	FunctionList []PHPFunction
}

// PHPFunction 表示一个php函数
type PHPFunction struct {
	FunctionName string
	Comment      string
}

// ParsePHP 解析php
func ParsePHP(f string) PHPClass {
	result := PHPClass{
		ClassName:    "",
		FunctionList: []PHPFunction{},
	}

	content := FileGetContents(f)
	lineList := strings.Split(content, "\n")

	for lineNum, line := range lineList {
		//fmt.Println(lineNum, line)

		line := strings.TrimSpace(line)

		//寻找class
		if result.ClassName == "" {
			if strings.Index(line, "class") == 0 {
				lineRes := strings.Split(line, " ")
				className := lineRes[1]
				if lastIndex := strings.LastIndex(className, "Action"); lastIndex > 0 {
					className = className[:lastIndex]
				} else if lastIndex := strings.LastIndex(className, "Controller"); lastIndex > 0 {
					className = className[:lastIndex]
				}

				result.ClassName = StrToLower(className)
			} else {
				continue
			}
		}

		//寻找function
		if strings.Contains(line, "function") {
			lineRes := strings.Split(line, " ")
			funcName := ""
			if lineRes[0] == "public" {
				funcName = lineRes[2]
			} else if lineRes[0] == "function" {
				funcName = lineRes[1]
			} else {
				continue
			}

			leftBracketIndex := strings.Index(funcName, "(")
			if leftBracketIndex > 0 {
				funcName = funcName[0:leftBracketIndex]
			}

			result.FunctionList = append(result.FunctionList, PHPFunction{
				FunctionName: funcName,
				Comment:      getPHPComment(lineNum, lineList),
			})
		}
	}
	//fmt.Println(JSONEncode(result))
	return result
}

func getPHPComment(lineNum int, lineList []string) string {
	findCommentLine := false

	lastCommentLine := -1
	firstCommentLine := -1
	for {
		lineNum--

		if lineNum < 0 {
			break
		}

		line := strings.TrimSpace(lineList[lineNum])
		if !findCommentLine && line == "}" {
			break
		}

		if StringLen(line) >= 2 {
			if lastCommentLine == -1 {
				if line[0:2] == "//" {
					lastCommentLine = lineNum
					findCommentLine = true
					continue
				}

				if line[StringLen(line)-2:] == "*/" {
					lastCommentLine = lineNum
					continue
				}
			}
			if findCommentLine {
				if line[0:2] == "//" {
					firstCommentLine = lineNum
					continue
				} else {
					break
				}
			} else {
				if line[0:2] == "/*" {
					firstCommentLine = lineNum
					break
				}
			}
		}
	}
	if firstCommentLine == -1 || lastCommentLine == -1 {
		return ""
	}
	res := lineList[firstCommentLine : lastCommentLine+1]

	var temp []string

	for _, line := range res {
		// startIndex := -1

		// for k, v := range line {
		// 	if !InStringSlice(strings.TrimSpace(line[k:k+len(string(v))]), []string{"", "*", "/"}) {
		// 		startIndex = k
		// 		break
		// 	}
		// }
		temp = append(temp, strings.TrimSpace(line))
	}

	return strings.Replace(Implode("\n", temp), "\r", "", -1)
}

// GetPHPConfigJSON 获取php配置
func GetPHPConfigJSON(path string) string {
	return ExecuteQuiet([]string{"php", "-r", "echo json_encode(include('" + path + "'),JSON_UNESCAPED_UNICODE| JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_LINE_TERMINATORS);"})
}
