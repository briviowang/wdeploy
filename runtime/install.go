package runtime

func InstallWindowsEnv() {
	Rm(GetRuntimeConfigDir() + "/runtime")

	for _, name := range []string{"git", "php", "putty"} {
		installGet(name)
	}
	StepLog("安装vcredist_vc2012_x86")
	Execute([]string{
		"vcredist_vc2012_x86",
		"/install",
		"/quiet",
	})
	Rm(GetRuntimeConfigDir() + "/runtime/php/vcredist_vc2012_x86.exe")
	Rm(GetRuntimeConfigDir() + "/temp")
}

func installGet(name string) {
	tempDir := GetRuntimeConfigDir() + "/temp/"

	if !FileExists(tempDir) {
		Mkdir(tempDir)
	}
	StepLog("安装" + name)

	url := "https://brivio-assets.oss-cn-hangzhou.aliyuncs.com/wdeploy-install/" + name + "-windows.zip"

	HTTPService := HTTPService{}
	HTTPService.Download(url, tempDir)

	Unzip(tempDir+"/"+name+"-windows.zip", GetRuntimeConfigDir()+"/runtime/"+name+"/")
}

func (c *WebDeployService) cmdInstall() {
	cmd := "install"
	c.AddHelp(cmd, `安装windows环境`)

	if argsAct != cmd {
		return
	}
	InstallWindowsEnv()

	RuntimeExit()
}
