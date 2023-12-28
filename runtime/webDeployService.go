package runtime

import (
	"fmt"
	"os"
	"strings"
)

// WebDeployUploadCacheItem 上传配置项
type WebDeployUploadCacheItem struct {
	LastUploadTime int64
	HashKey        string
	ServerHost     string
	ServerPath     string
}

// WebDeployCache 上传的服务器列表
type WebDeployCache struct {
	UploadList []WebDeployUploadCacheItem
}

// ArgTime time
var ArgTime = "time"

// WebDeployService 网站部署
type WebDeployService struct {
	BaseCommandService

	uploadZipName      string
	uploadZipPath      string
	uploadIgnoreFiles  []string
	garbageIgnoreFiles []string
	uploadConfigDir    string
	uploadConfigFile   string
	uploadCache        WebDeployCache
	webRoot            string

	uploadParams map[string]interface{}
}

// NewWebDeployService 实例化
func NewWebDeployService() *WebDeployService {
	c := WebDeployService{}
	c.init()
	return &c
}

var currentDeployConfig *WebDeployConfigItem

// WebDeployConfigItem 配置项
type WebDeployConfigItem struct {
	Name           string
	ServerHost     string
	ServerPort     string
	ServerUser     string
	ServerPath     string
	IgnoreFiles    []string
	SSHPrivateFile string
	SSHJumpHost    string
	//配置例子：ssh_proxy_command="ProxyCommand ssh root@gcp1 nc -w 1 %h %p"
	SSHProxyCommand     string
	EncryptPHPDirectory string
	EncryptExcludeDir   string
	AfterUploadScript   string
	XcxDir              string
	TpVersion           string
}

var globalConfig *WebDeployGlobalConfig

type WebDeployGlobalConfig struct {
	PHPPath    string
	GitPath    string
	XCXCliPath string
}

func initDeployConfig() {
	globalConfig = &WebDeployGlobalConfig{
		PHPPath:    "php",
		GitPath:    "git",
		XCXCliPath: "wxcli",
	}
	// 加载全局配置
	// globalConfigRes := LoadConfig(GetRuntimeConfigDir() + "/wdeploy.config")
	// globalConfig.PHPPath = assetsDefault([]string{getArg("php-path", ""), globalConfigRes["php_path"].String()}, "请配置PHP路径")
	// globalConfig.GitPath = assetsDefault([]string{getArg("git-path", ""), globalConfigRes["git_path"].String()}, "请配置Git路径")
	// globalConfig.XCXCliPath = getArg("xcx-cli-path", globalConfigRes["xcx_cli_path"].String())

	configFilePath := Pwd() + "/deploy.config"

	if hasArg("config") {
		configFilePath = argsParams["config"]
	}

	vars := LoadConfig(configFilePath)

	serverUser := getArg("server-user", vars["server_user"].String())
	if IsEmpty(Trim(serverUser)) {
		serverUser = "root"
	}

	serverHost := getArg("server-host", vars["hostname"].String())
	if IsEmpty(Trim(serverHost)) {
		ErrorLog("未定义HOSTNAME!")
	}

	serverPort := getArg("server-port", vars["server_port"].String())
	if IsEmpty(Trim(serverPort)) {
		serverPort = "22"
	}

	name := getArg("name", vars["name"].String())
	if IsEmpty(Trim(name)) {
		name = serverHost
	}

	currentDeployConfig = &WebDeployConfigItem{
		ServerUser:          serverUser,
		ServerHost:          serverHost,
		ServerPort:          serverPort,
		ServerPath:          getArg("server-path", vars["server_path"].String()),
		SSHPrivateFile:      getArg("ssh-private-file", vars["ssh_private_file"].String()),
		SSHJumpHost:         getArg("ssh-jump-host", vars["ssh_jump_host"].String()),
		SSHProxyCommand:     getArg("ssh-proxy-command", vars["ssh_proxy_command"].String()),
		EncryptPHPDirectory: getArg("encrypt-php-dir", vars["encrypt_php_dir"].String()),
		EncryptExcludeDir:   getArg("encrypt-exclude-dir", vars["encrypt_exclude_dir"].String()),
		AfterUploadScript:   getArg("after-upload-script", vars["after_upload_script"].String()),
		XcxDir:              getArg("xcx-dir", vars["xcx_dir"].String()),
		Name:                name,
	}
	ignoreFiles := TrimSpace(vars["ignore_files"].String())
	if StringLen(ignoreFiles) > 0 {
		currentDeployConfig.IgnoreFiles = strings.Split(ignoreFiles, " ")
	}

	logHint = fmt.Sprintf(" [%s:%s]", currentDeployConfig.ServerHost, currentDeployConfig.ServerPath)
	//tp版本
	if FileExists(Pwd() + "/data/config/db.php") {
		currentDeployConfig.TpVersion = "3"
	} else if FileExists(Pwd() + "/data/config/database.php") {
		currentDeployConfig.TpVersion = "5"
	} else if FileExists(Pwd() + "/config/database.php") {
		currentDeployConfig.TpVersion = "6"
	}
}

func (c *WebDeployService) init() {
	c.webRoot = Pwd()
	c.uploadParams = map[string]interface{}{}

	InitRuntime()

	if c.isWebRoot() {
		c.garbageIgnoreFiles = []string{
			"Thumbs.db",
			".DS_Store",
			"*.bat",
			"*.zip",
			"*.log",
			"*.bak",
			"*.less",
		}

		c.uploadIgnoreFiles = []string{
			"README.*",
			"/*.sql",
			"*.code-workspace",
			"/*api.config.php",
			"/*api.test.php",
			"/*Api.config.php",
			"/data/config/db.php",
			"/data/config/database.php",
			"/index.php",
			"/deploy",
			"/deploy.*",
			"/*.sh",
			"/compare.sql",
			"/data/upload/*",
			"/data/runtime/*",
			"*/node_modules/*",
			"/test.*",
			"/tools/sdk/*",
			"/tools/api.phar",
		}
		if len(currentDeployConfig.IgnoreFiles) > 0 {
			c.uploadIgnoreFiles = append(c.uploadIgnoreFiles, currentDeployConfig.IgnoreFiles...)
		}

		c.uploadZipName = "dist.zip"
		c.uploadZipPath = c.webRoot + "/" + c.uploadZipName
		if FileExists(c.uploadZipPath) {
			Rm(c.uploadZipPath)
		}

		//检查git
		gitDir := c.webRoot + "/.git"
		if !FileExists(gitDir) {
			ExecuteString(globalConfig.GitPath + " init")
		}

		if FileExists(gitDir) {
			c.uploadConfigDir = gitDir + "/deploy"
			if !FileExists(c.uploadConfigDir) {
				Mkdir(c.uploadConfigDir)
			}

			c.uploadConfigFile = c.uploadConfigDir + "/config.json"
			if !FileExists(c.uploadConfigFile) {
				FilePutContents(c.uploadConfigFile, "{}")
			}
			c.saveLastUploadTime(0)

			// merge后更新上传时间，避免上传其他人的代码
			postMergeHook := gitDir + "/hooks/post-merge"
			FilePutContents(postMergeHook, `#!/bin/sh
if type wdeploy &> /dev/null;then
	wdeploy reset-upload >/dev/null
else
	echo "wdeploy not found"
fi
			`)
			os.Chmod(postMergeHook, 0777)
		}

		ignoreFile := c.webRoot + "/.gitignore"
		if !FileExists(ignoreFile) {
			FilePutContents(ignoreFile, `
.*
*.code-workspace
!.gitignore
!.htaccess
Thumbs.db
node_modules
*- Copy.*
复件*
/docs
/misc
/release
/data/runtime
/runtime
/data/upload
/data/config/db.php
/data/config/database.php
/testApi.config.php
/static/build/*
/data/config/url.php
/tools/sdk
/test.*
dict.utf8.xdb
data/font
tools/task/vendor
/*.sql
`)
		}

		htaccessFile := c.webRoot + "/.htaccess"
		if currentDeployConfig.TpVersion == "6" {
			htaccessFile = c.webRoot + "/public/.htaccess"
		}
		if !FileExists(htaccessFile) {
			FilePutContents(htaccessFile, `
<IfModule mod_rewrite.c>
RewriteEngine on

# 图片404配置
RewriteCond %{DOCUMENT_ROOT}%{REQUEST_URI} !-f
RewriteRule \.(gif|jpe?g|png|bmp) data/upload/NOPIC.png [NC,L]

# thinkphp路由配置
RewriteCond %{REQUEST_FILENAME} !-d
RewriteCond %{REQUEST_FILENAME} !-f
RewriteRule ^(.*)$ index.php/$1 [QSA,PT,L]
</IfModule>
`)
		}
	}

	if c.isWebRoot() {
		c.cmdAPI()
		c.cmdLastFiles()
		c.cmdZip()
		c.cmdLastUpload()
		c.cmdExecute()
		c.cmdResetUpload()
		c.cmdUpload()
		c.cmdUploadAll()
		c.cmdXcxUpload()
		c.cmdDBInfo()
		c.cmdDBCompare()
		c.cmdDBCheck()
		c.cmdExampleConfig()
		c.cmdAliyun()
		c.cmdDoc()
		c.cmdDBDoc()
		c.cmdDBStatics()
	}

	if IsWindows() {
		c.cmdInstall()
	}

	c.ShowHelp()
}

func (c *WebDeployService) isWebRoot() bool {
	return FileExists(c.webRoot + "/deploy.config")
}

// cmdExecute 远程执行一些操作
// 用法：
//
//	wdeploy execute --action=lnmp 安装apache、php、mysql
//	wdeploy execute --action=update_mysql_conf 更新mysql配置
func (c *WebDeployService) cmdExecute() {
	cmd := "execute"
	c.AddHelp(cmd, `远程执行一些操作,lnmp、update_mysql_conf`)

	if argsAct != cmd {
		return
	}
	c.uploadParams["action"] = argsParams["action"]
	c.executeServerScript()

	RuntimeExit()
}

func (c *WebDeployService) cmdExampleConfig() {
	cmd := "example-config"
	c.AddHelp(cmd, `显示配置文件样例`)

	if argsAct != cmd {
		return
	}
	contentBytes, err := Asset("assets/scripts/example.config")
	CheckErr(err)
	println(string(contentBytes))

	RuntimeExit()
}
