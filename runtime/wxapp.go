package runtime

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"
)

// WXAppConfig 配置项
type WXAppConfig struct {
	Pages          []string                   `json:"pages,omitempty"`
	Window         *WxAppConfigWindow         `json:"window,omitempty"`
	TabBar         *WxAppConfigTabBar         `json:"tabBar,omitempty"`
	NetworkTimeout *WxAppConfigNetworkTimeout `json:"networkTimeout,omitempty"`
	Debug          bool                       `json:"debug,omitempty"`
}

// WxAppConfigWindow 配置项
type WxAppConfigWindow struct {
	NavigationBarBackgroundColor string `json:"navigationBarBackgroundColor,omitempty"`
	NavigationBarTextStyle       string `json:"navigationBarTextStyle,omitempty"`
	NavigationBarTitleText       string `json:"navigationBarTitleText,omitempty"`
	NavigationStyle              string `json:"navigationStyle,omitempty"`
	BackgroundColor              string `json:"backgroundColor,omitempty"`
	BackgroundTextStyle          string `json:"backgroundTextStyle,omitempty"`
	BackgroundColorTop           string `json:"backgroundColorTop,omitempty"`
	BackgroundColorBottom        string `json:"backgroundColorBottom,omitempty"`
	EnablePullDownRefresh        string `json:"enablePullDownRefresh,omitempty"`
	OnReachBottomDistance        int    `json:"onReachBottomDistance,omitempty"`
}

// WxAppConfigTabBar 配置项
type WxAppConfigTabBar struct {
	Color           string            `json:"color,omitempty"`
	SelectedColor   string            `json:"selectedColor,omitempty"`
	BackgroundColor string            `json:"backgroundColor,omitempty"`
	BorderStyle     string            `json:"borderStyle,omitempty"`
	List            []WxAppConfigList `json:"list,omitempty"`
	Position        string            `json:"position,omitempty"`
}

// WxAppConfigList 配置项
type WxAppConfigList struct {
	PagePath         string `json:"pagePath,omitempty"`
	Text             string `json:"text,omitempty"`
	IconPath         string `json:"iconPath,omitempty"`
	SelectedIconPath string `json:"selectedIconPath,omitempty"`
}

// WxAppConfigNetworkTimeout 配置项
type WxAppConfigNetworkTimeout struct {
	Request       int `json:"request,omitempty"`
	ConnectSocket int `json:"connectSocket,omitempty"`
	UploadFile    int `json:"uploadFile,omitempty"`
	DownloadFile  int `json:"downloadFile,omitempty"`
}

func parseWxAppConf(confPath string) {
	if !FileExists(confPath) {
		return
	}
	conf := WXAppConfig{}
	json.Unmarshal([]byte(FileGetContents(confPath)), &conf)

	indexPage := "pages/index/index"

	if conf.TabBar != nil {
		if len(conf.TabBar.List) > 0 {
			indexPage = conf.TabBar.List[0].PagePath
		}
	}

	pagesDir := Dir(confPath) + "/pages/"
	var pages []string

	files, err := ioutil.ReadDir(pagesDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if !f.Mode().IsRegular() {
			if InStringSlice(f.Name(), []string{"common"}) {
				continue
			}
			if BaseName(indexPage) == f.Name() {
				continue
			}
			if !FileExists(pagesDir + f.Name() + "/" + f.Name() + ".js") {
				continue
			}
			pages = append(pages, fmt.Sprintf("pages/%s/%s", f.Name(), f.Name()))
		}
	}
	sort.Strings(pages)
	conf.Pages = append([]string{indexPage}, pages...)
	FilePutContents(confPath, JSONEncode(conf))
}

func (c *WebDeployService) cmdXcxUpload() {
	xcxDir := Pwd() + "/" + currentDeployConfig.XcxDir

	if !FileExists(xcxDir) {
		ErrorLog(xcxDir + "路径不存在")
		return
	}

	cmd := "xcx-upload"
	c.AddHelp(cmd, `上传小程序`)

	if argsAct != cmd {
		return
	}
	// 生成版本号
	version := time.Unix(Time(), 0).Format("0601021504")
	FilePutContents(xcxDir+"/src/version.js", "module.exports.version = '"+version+"';")

	command := exec.Command("npm", "run", "build:mp-weixin")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = xcxDir
	command.Run()

	// parseWxAppConf(wxapp + "/app.json")

	cliPath := globalConfig.XCXCliPath
	if IsEmpty(cliPath) {
		ErrorLog("需要配置xcx_cli_path")
		RuntimeExit()
	}

	Execute([]string{
		cliPath,
		"upload",
		"--project", xcxDir + "/dist/build/mp-weixin",
		"-v", version,
		"-d", "版本更新",
	})
	RuntimeExit()
}
