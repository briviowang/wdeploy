package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *WebDeployService) saveLastUploadTime(time int64) {
	json.Unmarshal([]byte(FileGetContents(c.uploadConfigFile)), &c.uploadCache)

	hashKey := c.getUploadHashKey()
	isConfig := false

	for key, val := range c.uploadCache.UploadList {
		if val.HashKey == hashKey {
			isConfig = true
			if time > 0 {
				c.uploadCache.UploadList[key].LastUploadTime = time
			}
			break
		}
	}

	if !isConfig {
		item := WebDeployUploadCacheItem{
			ServerPath:     currentDeployConfig.ServerPath,
			ServerHost:     currentDeployConfig.ServerHost,
			HashKey:        hashKey,
			LastUploadTime: time,
		}
		c.uploadCache.UploadList = append(c.uploadCache.UploadList, item)
	}
	FilePutContents(c.uploadConfigFile, JSONEncode(c.uploadCache))
}

func (c *WebDeployService) getUploadHashKey() string {
	return Md5(currentDeployConfig.ServerHost + currentDeployConfig.ServerPath)
}

func (c *WebDeployService) getLastUploadTime() int64 {
	json.Unmarshal([]byte(FileGetContents(c.uploadConfigFile)), &c.uploadCache)

	hashKey := c.getUploadHashKey()

	for _, val := range c.uploadCache.UploadList {
		if val.HashKey == hashKey {
			return val.LastUploadTime
		}
	}
	return 0
}

func (c *WebDeployService) upload() {
	if !FileExists(c.uploadZipPath) {
		ErrorLog(c.uploadZipPath + "不存在!")
	} else {
		fmt.Println("压缩包大小:" + FormatSize(GetFileSize(c.uploadZipPath)))
	}

	StepLog("开始上传")

	ssh := initSSHService()

	ssh.Upload(c.uploadZipPath, currentDeployConfig.ServerPath)
	Rm(c.uploadZipPath)

	c.executeServerScript()
}

func (c *WebDeployService) executeServerScript() {
	ssh := initSSHService()
	StepLog("远程操作")
	outShellFile := c.webRoot + "/.git/deploy/server.sh"

	contentBytes, err := Asset("assets/scripts/server.sh")
	CheckErr(err)

	if c.uploadParams["action"] == nil {
		c.uploadParams["action"] = ""
	}

	c.uploadParams["ServerPath"] = currentDeployConfig.ServerPath
	c.uploadParams["logColor"] = logColor
	c.uploadParams["config"] = currentDeployConfig

	content := ParseTemplate(string(contentBytes), c.uploadParams)
	FilePutContents(outShellFile, content)

	ssh.Shell(content)
	// Rm(outShellFile)
}

func (c *WebDeployService) addUnstagedFiles() {
	if !FileExists(c.webRoot + "/.git") {
		return
	}
	content := strings.Split(ExecuteStringQuiet(globalConfig.GitPath+" status -s"), "\n")

	var files []string

	for _, val := range content {
		if len(val) <= 0 {
			continue
		}

		if val[0:2] == "??" {
			files = append(files, val[3:])

			fPath := c.webRoot + "/" + val[3:]
			os.Chtimes(fPath, time.Now(), time.Now())
		}
	}
	if len(files) > 0 {
		Execute(append([]string{
			globalConfig.GitPath,
			"add",
		}, files...))
	}
}

func (c *WebDeployService) getLastEditFiles() []string {
	//根据git提交日期筛选文件
	if val, ok := argsParams["after"]; ok {
		content := ExecuteQuiet([]string{
			globalConfig.GitPath,
			"log",
			"master",
			fmt.Sprintf("--after='%v'", val),
			"--name-only",
			"--oneline",
		})
		result := []string{}
		res := uniqueStrings(strings.Split(content, "\n"))
		for _, val := range res {
			if len(val) == 0 {
				continue
			}
			val = c.webRoot + "/" + val
			if FileExists(val) {
				result = append(result, val)
			}

		}
		return result
	}

	c.addUnstagedFiles()
	t := time.Unix(c.getLastUploadTime(), 0)

	if hasArg(ArgTime) {
		unit := argsParams[ArgTime][len(argsParams[ArgTime])-1:]
		unitVal := argsParams[ArgTime][:len(argsParams[ArgTime])-1]
		offsetTime := 0
		if unit == "h" {
			offsetTime = StringToInt(unitVal) * 3600
		} else if unit == "m" {
			offsetTime = StringToInt(unitVal) * 60
		} else if unit == "d" {
			offsetTime = StringToInt(unitVal) * 3600 * 24
		} else {
			offsetTime = StringToInt(unitVal)
		}
		t = time.Unix(Time()-int64(offsetTime), 0)
	}

	find := FindService{
		Path:       c.webRoot,
		ModifyTime: t,
		Rules:      c.uploadIgnoreFiles,
	}
	return find.Find()
}

func checkLineEnding(file string) {
	if StrToLower(filepath.Ext(file)) == ".sh" {
		FilePutContents(file, strings.Replace(FileGetContents(file), "\r\n", "\n", -1))
	}
}

func (c *WebDeployService) zipDist(files []string) {
	if len(files) < 50 {
		fmt.Println(Implode("\n", TrimFilesPath(files, c.webRoot)))
	}
	fmt.Printf("文件数量:%d  ", len(files))

	// for _, file := range files {
	// 	checkLineEnding(file)
	// }

	Zip(c.uploadZipPath, files, c.webRoot)
}

func (c *WebDeployService) lastZip() int {
	files := c.getLastEditFiles()

	if len(os.Args) >= 3 {
		if os.Args[2] == "all" {
			content := strings.Split(ExecuteStringQuiet(globalConfig.GitPath+" status -s"), "\n")

			for _, val := range content {
				if len(val) <= 0 {
					continue
				}
				files = append(files, c.webRoot+"/"+TrimSpace(val[2:]))
			}
		}
	}
	if len(files) > 0 {
		c.zipDist(files)
		return len(files)
	}

	return 0
}

func (c *WebDeployService) cmdZip() {
	cmd := "zip"
	c.AddHelp(cmd, `打包最近修改的,可传参数all`)
	if argsAct != cmd {
		return
	}

	if c.lastZip() == 0 {
		SuccessLog("检查完毕，没有文件需要打包")
	} else {
		println("")
	}
	RuntimeExit()
}

func (c *WebDeployService) cmdLastUpload() {
	cmd := "last-upload"
	c.AddHelp(cmd, `上传最近修改的,可传参数all:重新全部上传`)
	if argsAct != cmd {
		return
	}

	if c.lastZip() > 0 {
		c.upload()
		c.saveLastUploadTime(Time())
	} else {
		SuccessLog("检查完毕，没有文件需要上传")
	}

	RuntimeExit()
}

func (c *WebDeployService) cmdLastFiles() {
	cmd := "last-files"
	c.AddHelp(cmd, `显示将要上传的文件`)
	if argsAct != cmd {
		return
	}
	files := c.getLastEditFiles()
	if len(files) > 0 {
		if len(files) > 50 {
			println(Implode("\n", TrimFilesPath(files[:49], c.webRoot)))
			println("...")
		} else {
			println(Implode("\n", TrimFilesPath(files, c.webRoot)))
		}
	} else {
		SuccessLog("检查完毕，没有文件需要上传")
	}

	RuntimeExit()
}

func (c *WebDeployService) cmdResetUpload() {
	cmd := "reset-upload"
	c.AddHelp(cmd, `重制上传时间`)
	if argsAct != cmd {
		return
	}
	c.saveLastUploadTime(Time())
	RuntimeExit()
}

func (c *WebDeployService) cmdUpload() {
	cmd := "upload"
	c.AddHelp(cmd, `上传文件到远程服务器，默认只上传代码`)
	if argsAct != cmd {
		return
	}

	if len(os.Args) >= 3 {
		var path string
		if StrPos(os.Args[2], c.webRoot) >= 0 {
			path = os.Args[2]
		} else {
			path = c.webRoot + "/" + strings.Trim(os.Args[2], "/\\")
		}
		if !FileExists(path) || len(path) == 0 {
			ErrorLog(path + "不存在")
		}

		c.zipDist([]string{path})

		c.upload()
	}
	RuntimeExit()
}

func (c *WebDeployService) cmdUploadAll() {
	cmd := "upload-all"
	c.AddHelp(cmd, `上传全部代码、图片到远程服务器
注意：该命令会预先把远程_core、app、static目录删除，再解压上传的代码`)

	if argsAct != cmd {
		return
	}
	find := FindService{
		Path:  c.webRoot,
		Rules: c.uploadIgnoreFiles,
	}
	files := find.Find()

	extraFiles := []string{
		"/index.php",
	}
	for _, f := range extraFiles {
		f = c.webRoot + f
		if FileExists(f) {
			files = append(files, f)
		}
	}
	if len(files) > 0 {
		c.zipDist(files)
		c.uploadParams["action"] = "remove"
		c.upload()
		c.saveLastUploadTime(Time())
	} else {
		SuccessLog("检查完毕，没有文件需要上传")
	}

	RuntimeExit()
}
