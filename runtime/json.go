package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

// JSONReader 读取json
type JSONReader struct {
}

// NewJSONReader 实例化
func NewJSONReader() *JSONReader {
	return &JSONReader{}
}

func (c *JSONReader) isEndingChar(char string) bool {
	return InStringSlice(char, []string{"\r", "\n"})
}

func (c *JSONReader) isBlockStartTag(char string) bool {
	return InStringSlice(char, []string{"[", "{"})
}

func (c *JSONReader) isBlockCloseTag(char string) bool {
	return InStringSlice(char, []string{"]", "}"})
}

func (c *JSONReader) isTag(char string) bool {
	return InStringSlice(char, []string{"[", "{", "]", "}", ",", "\""})
}

func (c *JSONReader) getPrevVisibleChar(pos int, s string) string {
	char := ""
	for i := pos - 1; i >= 0; i-- {
		char = Trim(string(s[i]))
		if char != "" {
			break
		}
	}
	return char
}

func (c *JSONReader) getNextVisibleChar(pos int, s string) string {
	maxLen := StringLen(s)
	char := ""
	for i := pos + 1; i < maxLen-1; i++ {
		char = Trim(string(s[i]))
		if char != "" {
			break
		}
	}
	return char
}

func (c *JSONReader) Read(data string) string {
	result := ""
	isInString := false
	isLineComment := false
	isBlockComment := false

	maxLen := StringLen(data)

	for k, v := range data {
		prevChar := ""
		nextChar := ""

		prevVisibleChar := c.getPrevVisibleChar(k, data)
		nextVisibleChar := c.getNextVisibleChar(k, data)

		if k > 0 {
			prevChar = TrimSpace(string(data[k-1]))
		}

		if k < maxLen-1 {
			nextChar = TrimSpace(string(data[k+1]))
		}

		char := TrimSpace(string(v))

		if char == "'" && !isInString {
			char = "\""
		}

		if char == "\"" && !isLineComment && !isBlockComment {
			isInString = !isInString
		}

		if !isInString {
			if char == "" {
				continue
			}
			if char == "/" {
				if !isLineComment && !isBlockComment {
					if nextChar == "/" {
						isLineComment = true
					}
					if nextChar == "*" {
						isBlockComment = true
					}
				}
				if isBlockComment && prevChar == "*" {
					isBlockComment = false
					continue
				}
			}

			if c.isEndingChar(char) && isLineComment {
				isLineComment = false
				continue
			}

			if isLineComment || isBlockComment {
				continue
			}
			if !isInString && c.isEndingChar(char) {
				continue
			}

			if char == "," && c.isBlockCloseTag(nextVisibleChar) {
				continue
			}
			if !c.isTag(char) && c.isTag(prevVisibleChar) && prevVisibleChar != "\"" {
				result += "\""
			}
			if char == ":" && !c.isTag(prevVisibleChar) {
				result += "\""
			}
		}

		result += char
	}

	return result
}

// JSONService 解析json类
type JSONService struct {
	Data []byte
	File string
}

func (c *JSONService) init() *JSONService {

	if c.Data == nil {
		c.Data = []byte(NewJSONReader().Read(FileGetContents(c.File)))
	}
	return c
}

// GetString 获取字符串
func (c *JSONService) GetString(keys ...string) string {
	c.init()
	result, _ := jsonparser.GetString(c.Data, keys...)
	return result
}

// GetInt 获取数字
func (c *JSONService) GetInt(keys ...string) int64 {
	c.init()
	result, _ := jsonparser.GetInt(c.Data, keys...)
	return result
}

// GetBoolean 获取bool
func (c *JSONService) GetBoolean(keys ...string) bool {
	c.init()
	result, _ := jsonparser.GetBoolean(c.Data, keys...)
	return result
}

// Each 数组列举
func (c *JSONService) Each(callback func(key string, value string, dataType jsonparser.ValueType), keys ...string) {
	c.init()
	_, dataType, _, _ := jsonparser.Get(c.Data, keys...)

	if dataType == jsonparser.Array {
		index := 0
		jsonparser.ArrayEach(c.Data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			callback(fmt.Sprintf("%d", index), string(value), dataType)
			index++
		}, keys...)
	} else if dataType == jsonparser.Object {
		jsonparser.ObjectEach(c.Data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			callback(string(key), string(value), dataType)
			return nil
		}, keys...)
	}
}

// GetType 获取类型
func (c *JSONService) GetType(keys ...string) jsonparser.ValueType {
	c.init()
	_, dataType, _, _ := jsonparser.Get(c.Data, keys...)
	return dataType
}

// Set 设置键值
func (c *JSONService) Set(value string, keys ...string) {
	c.init()
	jsonparser.Set(c.Data, []byte(value), keys...)
}

// JSONEncode 转换成json字符串
func JSONEncode(d interface{}) string {
	result, _ := json.MarshalIndent(d, "", "    ")
	return string(result)
}

// JSONConvert struct转换
func JSONConvert(from interface{}, to interface{}) {
	byteData, _ := json.Marshal(from)
	json.Unmarshal(byteData, to)
}
