package runtime

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/shell"
)

var (
	argsAct    string
	argsParams map[string]string
)

func IsDebug() bool {
	if val, ok := argsParams["foo"]; ok {
		return val != "no"
	}
	return false
}

// IsWindows 是否是windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsDarwin 是否是mac os
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// ExecutePHP 执行php
func ExecutePHP(name string, data map[string]interface{}) string {
	outPHPFile := Pwd() + "/.git/deploy/execute.php"

	contentBytes, err := Asset("assets/scripts/" + name + ".php")
	CheckErr(err)

	content := ParseTemplate(string(contentBytes), data)
	FilePutContents(outPHPFile, content)

	result := ExecuteQuiet([]string{
		globalConfig.PHPPath,
		outPHPFile,
	})

	// Rm(outPHPFile)

	return result
}

// ExecuteString 执行本地shell
func ExecuteString(shell string) {
	Execute(strings.Split(shell, " "))
}

// ExecuteStringQuiet 静默执行本地shell
func ExecuteStringQuiet(shell string) string {
	return ExecuteQuiet(strings.Split(shell, " "))
}

func createExecuteCmd(params []string) *exec.Cmd {
	name := params[0]

	cmd := exec.Command(name, params[1:]...)
	cmd.Dir = Pwd()

	return cmd
}

// Execute 执行本地shell
func Execute(params []string) error {
	var cmd *exec.Cmd

	if len(params) == 0 {
		return nil
	}
	cmd = createExecuteCmd(params)

	// if IsWindows() {
	// 	var handler *ExecuteHandler
	// 	cmd.Stdout = handler
	// }
	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecuteStart 执行shell
func ExecuteStart(params []string) *exec.Cmd {
	var cmd *exec.Cmd
	if len(params) == 0 {
		return cmd
	}
	cmd = createExecuteCmd(params)
	if IsDebug() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	InfoLog(strings.ReplaceAll(cmd.String(), "\\", "/"))
	cmd.Start()
	return cmd
}

// ExecuteQuiet 静默执行shell
func ExecuteQuiet(params []string) string {
	if len(params) == 0 {
		return ""
	}
	cmd := createExecuteCmd(params)
	cmd.Stderr = os.Stderr
	content, _ := cmd.Output()
	return strings.TrimSpace(string(content))
}
func formatPath(s string) string {
	//if IsWindows() {
	s = strings.Replace(s, "\\", "/", -1)
	//}
	return s
}

// Pwd 获取当前路径
func Pwd() string {
	dir, err := os.Getwd()
	CheckErr(err)
	return formatPath(dir)
}

// CheckErr 检查错误err
func CheckErr(err error) bool {
	if err != nil {
		log.Panic(err)
		return false
	}
	return true
}

// GetHomeDir 获取home路径
func GetHomeDir() string {
	usr, err := user.Current()
	CheckErr(err)

	return strings.ReplaceAll(usr.HomeDir, "\\", "/")
}

// GetExecuteDir 获取程序所在路径
func GetExecuteDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

// Time 获取
func Time() int64 {
	return time.Now().Unix()
}

// Date 获取格式化日期
func Date(t int64) string {
	return time.Unix(t, 0).Format(DATE_FORMAT)
}

// ParseArgs 解析参数
func ParseArgs() {
	argsAct = "help"
	argsParams = map[string]string{}

	for k, v := range os.Args {
		if k == 0 {
			continue
		}

		v = strings.TrimLeft(v, "-")

		res := strings.Split(v, "=")

		if k == 1 {
			argsAct = StrToLower(v)
		} else {
			if len(res) > 1 {
				argsParams[StrToLower(res[0])] = StrToLower(res[1])
			} else {
				argsParams[StrToLower(res[0])] = ""
			}

		}
	}
}

// hasArg 是否有参数
func hasArg(name string) bool {
	_, exist := argsParams[name]
	return exist
}

// GetArg 获取参数
func getArg(name string, defaultVal string) string {
	if hasArg(name) {
		return argsParams[name]
	}
	return defaultVal
}

// assetsDefault 校验values中是否有值，否则直接报错,例子:assetsDefault([]string{globalConfig.PHPPath}, "请配置php路径")
func assetsDefault(values []string, msg string) string {
	if len(values) == 0 {
		log.Panic(msg)
	}
	// values循环
	for _, v := range values {
		if IsEmpty(v) {
			continue
		}
		return v
	}
	log.Panic(msg)
	return ""
}

// BaseCommandService 帮助内容生成类
type BaseCommandService struct {
	HelpMap map[string]string
}

// AddHelp 添加help
func (c *BaseCommandService) AddHelp(cmd string, help string) {
	if c.HelpMap == nil {
		c.HelpMap = map[string]string{}
	}

	c.HelpMap[cmd] = help
}

// ShowHelp 显示help
func (c *BaseCommandService) ShowHelp() {
	if c.HelpMap == nil {
		c.HelpMap = map[string]string{}
	}

	cmd := "help"
	c.HelpMap[cmd] = `显示帮助`

	maxKeyLength := 0

	var keys []string
	for k := range c.HelpMap {
		keys = append(keys, k)
		if len(k) > maxKeyLength {
			maxKeyLength = len(k)
		}
	}

	sort.Strings(keys)

	fmt.Println("用法：")
	for _, k := range keys {
		ColorStepText("  " + k)
		fmt.Println("    " + strings.Replace(c.HelpMap[k], "\n", "\n    ", -1))
	}
	RuntimeExit()
}

// IsPortCanUse 端口是否占用
func IsPortCanUse(port int) bool {
	l, err := net.Listen("tcp", "127.0.0.1:"+IntToString(port))
	if err != nil {
		return false
	}

	defer l.Close()
	return true
}

// GetMapKeys 获取map的key列表
func GetMapKeys(m map[string]interface{}) []string {
	result := make([]string, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// GetExecutePath 获取命令所在路径
func GetExecutePath(name string) string {
	result, err := exec.LookPath(name)
	if err == nil {
		return result
	}
	return ""
}

// CommandExist 命令是否存在
func CommandExist(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetRuntimeConfigDir 获取程序缓存路径
func GetRuntimeConfigDir() string {
	return GetHomeDir() + "/.web-deploy"
}

// BaseName 获取路径上的文件名或文件夹名
func BaseName(p string) string {
	return filepath.Base(p)
}

// Dir 获取完整路径
func Dir(path string) string {
	return filepath.Dir(path)
}

// FileGetContents 获取文件内容
func FileGetContents(fileName string) string {
	content, error := ioutil.ReadFile(fileName)
	if error != nil {
		log.Println(error)
		return ""
	}
	return string(content)
}

// FilePutContents 写入文件内容
func FilePutContents(fileName string, content string) {
	paths, _ := filepath.Split(fileName)
	Mkdir(paths)

	ioutil.WriteFile(fileName, []byte(content), os.ModePerm)
}

// GetFileSize 获取文件大小
func GetFileSize(fPath string) int64 {
	file, err := os.Open(fPath)
	defer file.Close()

	if err != nil {
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return 0
	}

	return fileInfo.Size()
}

// GetFileModTime 获取文件修改时间
func GetFileModTime(s string) time.Time {
	f, _ := os.Open(s)
	defer f.Close()
	fInfo, _ := f.Stat()
	return fInfo.ModTime()
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(f string) string {
	res := strings.Split(StrToLower(f), ".")
	return res[len(res)-1]
}

// Mkdir 创建文件夹
func Mkdir(p string) {
	if !FileExists(p) {
		err := os.MkdirAll(p, 0777)

		if err != nil {
			log.Println("Error creating directory")
			log.Println(err)
		}
	}
}

// FileExists 文件是否存在
func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Rm 删除文件、文件夹
func Rm(p string) {
	if !FileExists(p) {
		return
	}

	filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			os.Remove(path)
		}
		return nil
	})
	err := os.RemoveAll(p)
	if err != nil {
		println(err.Error())
	}
}

// DATE_FORMAT 日期格式化
var DATE_FORMAT = "2006-01-02 15:04:05"

// IntToString int转换成string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// StringToInt string转换成int
func StringToInt(i string) int {
	result, _ := strconv.Atoi(i)
	return result
}

// Int64ToString int64转换成string
func Int64ToString(i int64) string {
	return strconv.FormatInt(Time(), 10)
}

// StringToInt64 string转换成int64
func StringToInt64(i string) int64 {
	result, _ := strconv.ParseInt(i, 10, 64)
	return result
}

// Ucfirst 首字母大写
func Ucfirst(s string) string {
	return strings.Title(s)
}

// StrPos 查找字符串位置
func StrPos(haystack string, needle string) int {
	return strings.Index(haystack, needle)
}

// StrToLower 转成小写
func StrToLower(s string) string {
	return strings.ToLower(s)
}

// StrToUpper 转换成大写
func StrToUpper(s string) string {
	return strings.ToUpper(s)
}

// StringLen 获取字符串长度
func StringLen(s string) int {
	return utf8.RuneCountInString(s)
}

// Intval 字符串转成int
func Intval(s interface{}) int {
	result, _ := strconv.Atoi(s.(string))
	return result
}

// Ceil 类似php Ceil
func Ceil(f float64) int {
	result := int(f)
	if float64(result) < f {
		result++
	}
	return result
}

// TrimChar 删除首尾特定字符串
func TrimChar(s, trim string) string {
	return strings.Trim(s, trim)
}

// TrimSpace 删除首尾空格
func TrimSpace(s string) string {
	s = strings.Trim(s, " ")
	return s
}

// Trim 删除首尾空格、换行
func Trim(s string) string {
	s = strings.Trim(s, " \r\n")
	return s
}

// IsEmpty 是否为空值
func IsEmpty(s interface{}) bool {
	switch s.(type) {
	case string:
		if s == nil {
			return true
		} else if s == "" {
			return true
		}
	case int64:
		if s == 0 {
			return true
		}
	}
	return false
}

// IsNotEmpty 是否非空
func IsNotEmpty(s interface{}) bool {
	return !IsEmpty(s)
}

// InStringSlice 是否在slice中
func InStringSlice(search string, s []string) bool {
	for _, v := range s {
		if v == search {
			return true
		}
	}
	return false
}

// InIntSlice 去重string slice
func uniqueStrings(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Md5 获取md5值
func Md5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// FormatSize 大小格式化
func FormatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/(1024*1024))
	}
	if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fGB", float64(size)/(1024*1024*1024))
	}
	return fmt.Sprintf("%.2fTB", float64(size)/(1024*1024*1024*1024))
}

// KeyExists 键值是否存在
func KeyExists(k string, m map[string]interface{}) bool {
	_, err := m[k]
	return err
}

// Explode 类似php Explode
func Explode(sep string, s string) []string {
	return strings.Split(s, sep)
}

// Implode 类似php Implode
func Implode(sep string, l []string) string {
	return strings.Join(l, sep)
}

// IsNumber 是否是数字
func IsNumber(i interface{}) bool {
	var s string
	switch i.(type) {
	case string:
		s = i.(string)
	case []byte:
		s = string(i.([]byte))
	}

	_, err := strconv.Atoi(s)
	return err == nil
}

// SliceRemove 删除slice中的某一项
func SliceRemove(data []interface{}, i int) []interface{} {
	return append(data[:i], data[i+1:]...)
}

// SliceReverse slice反转
func SliceReverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// StrToTime 字符串转时间戳
func StrToTime(date string) int64 {
	t, err := time.Parse(DATE_FORMAT, date)
	if err == nil {
		return t.Unix()
	}
	return 0
}

// GbkToUtf8 gbk转utf8
func GbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil
	}
	return d
}

// GbkToUtf8String 字符串gbk转utf8
func GbkToUtf8String(s string) string {
	return string(GbkToUtf8([]byte(s)))
}

// Utf8ToGbk utf8转gbk
func Utf8ToGbk(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil
	}
	return d
}

// Utf8ToGbkString utf8转gbk
func Utf8ToGbkString(s string) string {
	return string(Utf8ToGbk([]byte(s)))
}

// IsURL 是否是URL
func IsURL(s interface{}) bool {
	if IsEmpty(s) {
		return false
	}

	var u *url.URL
	switch s.(type) {
	case string:
		u, _ = url.Parse(s.(string))
	case url.URL:
		t := s.(url.URL)
		u = &t
	case *url.URL:
		u = s.(*url.URL)
	default:
		return false
	}

	return InStringSlice(u.Scheme, []string{"http", "https"})
}

func LoadConfig(configPath string) map[string]expand.Variable {
	vars := map[string]expand.Variable{}

	if FileExists(configPath) {
		res, err := shell.SourceFile(context.TODO(), configPath)
		if err != nil {
			println(err.Error())
		} else {
			for key, val := range res {
				vars[StrToLower(key)] = val
			}
		}
	}

	return vars
}

var runtimeBinList []string

// InitRuntime runtime初始化
func InitRuntime() {
	ParseArgs()
	InitLog()
	initDeployConfig()
	if !FileExists(GetRuntimeConfigDir()) {
		Mkdir(GetRuntimeConfigDir())
	}
}

// 驼峰转蛇形 snake string
func Camel2Snake(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90（A-Z），97-122（a-z）
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	//ToLower把大写字母统一转小写
	return strings.ToLower(string(data[:]))
}

// 蛇形转驼峰
func Snake2Camel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}
func Snake2CamelVar(s string) string {
	s = Snake2Camel(s)
	return strings.ToLower(s[:1]) + s[1:]
}

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func GetRandomPort() int {
	for port := 20000; port < 255*255; port++ {
		if IsPortCanUse(port) {
			return port
		}
	}
	return 0
}

func TrimFilesPath(files []string, prefix string) []string {
	var result []string
	for _, f := range files {
		if !FileExists(f) {
			continue
		}

		if strings.Index(f, prefix) == 0 {
			result = append(result, f[len(prefix):])
		}
	}
	return result
}
